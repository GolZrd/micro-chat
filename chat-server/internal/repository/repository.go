package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository interface {
	Create(ctx context.Context, dto CreateChatDTO) (int64, error)
	Delete(ctx context.Context, id int64) error
	SendMessage(ctx context.Context, msg MessageCreateDTO) error
	ChatExists(ctx context.Context, id int64) (bool, error)
	IsUserInChat(ctx context.Context, chatId, userId int64) (bool, error)
	RecentMessages(ctx context.Context, chatId int64, limit int) ([]MessageDTO, error)
	UserChats(ctx context.Context, userId int64) ([]ChatInfoDTO, error)
	FindDirectChat(ctx context.Context, userId1 int64, userId2 int64) (int64, error)
	CreateDirectChat(ctx context.Context, userId1 int64, userId2 int64, username1 string, username2 string) (int64, error)
}

type repo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) ChatRepository {
	return &repo{
		db: db,
	}
}

func (r *repo) Create(ctx context.Context, dto CreateChatDTO) (int64, error) {
	var chatId int64
	// Здесь используем простой запрос
	chatBuilder := squirrel.Insert("chats").
		PlaceholderFormat(squirrel.Dollar).
		Columns("name", "is_direct").
		Values(dto.Name, !dto.IsGroup).
		Suffix("RETURNING id")

	chatQuery, args, err := chatBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build chat insert query: %w", err)
	}

	err = r.db.QueryRow(ctx, chatQuery, args...).Scan(&chatId)
	if err != nil {
		return 0, fmt.Errorf("insert chat: %w", err)
	}

	// Теперь нужно создать записи в таблице chat_members, в которых будем указывать id только что созданного чата
	insertMembers := squirrel.Insert("chat_members").Columns("chat_id", "user_id", "username", "role")
	for i, member := range dto.Members {
		role := "member"
		if i == 0 && dto.IsGroup {
			role = "owner" // Первый пользователь в группе будет владельцем
		}

		insertMembers = insertMembers.Values(chatId, member.UserId, member.Username, role)

		query, args, err := insertMembers.ToSql()
		if err != nil {
			return 0, fmt.Errorf("build members insert query: %w", err)
		}

		_, err = r.db.Exec(ctx, query, args...)
		if err != nil {
			return 0, fmt.Errorf("insert chat members: %w", err)
		}
	}

	return chatId, nil
}

func (r *repo) Delete(ctx context.Context, chat_id int64) error {
	builder := squirrel.Delete("chats").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": chat_id})

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build delete query: %w", err)
	}

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete chat: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("chat not found")
	}

	return nil
}

func (r *repo) SendMessage(ctx context.Context, msg MessageCreateDTO) error {
	// Выполняем основной запрос, добавляем в таблицу messages
	builder := squirrel.Insert("messages").
		PlaceholderFormat(squirrel.Dollar).
		Columns("chat_Id", "user_id", "from_username", "text").
		Values(msg.ChatId, msg.UserId, msg.FromUsername, msg.Text)
	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build send message query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	// Обновляем updated_at в чате после отправки сообщения
	updateBuilder := squirrel.Update("chats").Set("updated_at", squirrel.Expr("NOW()")).Where(squirrel.Eq{"ID": msg.ChatId})
	updateQuery, updateArgs, err := updateBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build update chat query: %w", err)
	}

	_, err = r.db.Exec(ctx, updateQuery, updateArgs...)
	if err != nil {
		return fmt.Errorf("update chat: %w", err)
	}

	return nil
}

func (r *repo) RecentMessages(ctx context.Context, chatID int64, limit int) ([]MessageDTO, error) {
	builder := squirrel.Select("id", "chat_id", "user_id", "from_username", "text", "created_at").
		PlaceholderFormat(squirrel.Dollar).
		From("messages").
		Where(squirrel.Eq{"chat_id": chatID}).
		Limit(uint64(limit))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build recent messages query: %w", err)
	}

	var messages []MessageDTO
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}

	for rows.Next() {
		var message MessageDTO
		err := rows.Scan(&message.Id, &message.ChatId, &message.UserId, &message.From, &message.Text, &message.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func (r *repo) ChatExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	// Здесь используем простой запрос
	query := "SELECT EXISTS (SELECT 1 FROM chats WHERE id = $1)"
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check chat exists: %w", err)
	}
	return exists, nil
}

// Проверяем, что пользователь является участником чата
func (r *repo) IsUserInChat(ctx context.Context, chatId, userId int64) (bool, error) {
	builder := squirrel.Select("1").
		PlaceholderFormat(squirrel.Dollar).
		From("chat_members").
		Where(squirrel.Eq{"chat_id": chatId, "user_id": userId}).
		Limit(1)

	query, args, err := builder.ToSql()
	if err != nil {
		return false, fmt.Errorf("build query: %w", err)
	}

	var isMember bool
	err = r.db.QueryRow(ctx, query, args...).Scan(&isMember)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check chat membership: %w", err)
	}

	return true, nil
}

func (r *repo) UserChats(ctx context.Context, userId int64) ([]ChatInfoDTO, error) {
	builder := squirrel.Select("c.ID", "c.name", "c.is_direct", "c.created_at", "c.updated_at").
		PlaceholderFormat(squirrel.Dollar).
		From("chats c").
		Join("chat_members ON c.ID = chat_members.chat_id ").
		Where(squirrel.Eq{"chat_members.user_id": userId}).
		OrderBy("c.updated_at DESC")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query user chats: %w", err)
	}

	var chats []ChatInfoDTO
	for rows.Next() {
		var chat ChatInfoDTO
		err := rows.Scan(&chat.ID, &chat.Name, &chat.IsDirect, &chat.CreatedAt, &chat.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan chat: %w", err)
		}

		chats = append(chats, chat)
	}

	for i := range chats {
		members, err := r.chatMembers(ctx, chats[i].ID)
		if err != nil {
			return nil, fmt.Errorf("get chat members: %w", err)
		}
		chats[i].Members = members
	}

	return chats, nil
}

// Получаем участников чата
func (r *repo) chatMembers(ctx context.Context, chatID int64) ([]MemberDTO, error) {
	builder := squirrel.Select("user_id", "username").
		PlaceholderFormat(squirrel.Dollar).
		From("chat_members").
		Where(squirrel.Eq{"chat_id": chatID}).
		OrderBy("joined_at ASC")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query chat members: %w", err)
	}
	defer rows.Close()

	var members []MemberDTO
	for rows.Next() {
		var member MemberDTO
		err := rows.Scan(&member.UserId, &member.Username)
		if err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, member)
	}

	return members, nil
}

// FindDirectChat - находим личный чат между двумя пользователями
func (r *repo) FindDirectChat(ctx context.Context, userId1 int64, userId2 int64) (int64, error) {
	// Личный чат — это чат ровно с 2 участниками
	builder := squirrel.Select("cm1.chat_id").
		PlaceholderFormat(squirrel.Dollar).
		From("chat_members cm1").
		Join("chats on chats.id = chat_members.chat_id").
		Join("chat_members cm2 on cm1.chat_id = cm2.chat_id").
		Where(squirrel.And{
			squirrel.Eq{"chats.is_direct": true},
			squirrel.Eq{"cm1.user_id": userId1},
			squirrel.Eq{"cm2.user_id": userId2},
		}).
		Limit(1)

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build query: %w", err)
	}

	var chatId int64
	err = r.db.QueryRow(ctx, query, args...).Scan(&chatId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("query direct chat: %w", err)
	}

	return chatId, nil
}

// CreateDirectChat - создаем личный чат между двумя пользователями
func (r *repo) CreateDirectChat(ctx context.Context, userId1 int64, userId2 int64, username1 string, username2 string) (int64, error) {
	// Сначала создаем сам чат
	builder := squirrel.Insert("chats").
		PlaceholderFormat(squirrel.Dollar).
		Columns("name", "is_direct").
		Values("", true).
		Suffix("RETURNING id")

	chatQuery, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build query: %w", err)
	}

	var chatId int64
	err = r.db.QueryRow(ctx, chatQuery, args...).Scan(&chatId)
	if err != nil {
		return 0, fmt.Errorf("create direct chat: %w", err)
	}

	// Теперь добавляем обоих участников в таблицу chat_members
	membersBuilder := squirrel.Insert("chat_members").
		Columns("chat_id", "user_id", "username", "role").
		Values(chatId, userId1, username1, "member").
		Values(chatId, userId2, username2, "member")

	membersQuery, args, err := membersBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build query: %w", err)
	}

	_, err = r.db.Exec(ctx, membersQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("add members to direct chat: %w", err)
	}

	return chatId, nil
}
