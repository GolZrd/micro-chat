package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository interface {
	Create(ctx context.Context, usernames []string) (int64, error)
	Delete(ctx context.Context, id int64) error
	SendMessage(ctx context.Context, msg MessageCreateDTO) error
	ChatExists(ctx context.Context, id int64) (bool, error)
	RecentMessages(ctx context.Context, chatID int64, limit int) ([]MessageDTO, error)
	UserChats(ctx context.Context, username string) ([]ChatInfoDTO, error)
}

type repo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) ChatRepository {
	return &repo{
		db: db,
	}
}

func (r *repo) Create(ctx context.Context, usernames []string) (int64, error) {
	var chatId int64
	// Здесь используем простой запрос, потому что у нас для chats по умолчанию дефолтные значения, поэтому не нужно дополнительно что то вставлять
	sql := "INSERT INTO chats DEFAULT VALUES RETURNING id"
	err := r.db.QueryRow(ctx, sql).Scan(&chatId)
	if err != nil {
		return 0, fmt.Errorf("insert chat: %w", err)
	}

	// Теперь нужно создать записи в таблице chat_members, в которых будем указывать id только что созданного чата
	for _, username := range usernames {
		inserMembers := squirrel.Insert("chat_members").
			PlaceholderFormat(squirrel.Dollar).
			Columns("chat_id", "username", "joined_at").
			Values(chatId, username, time.Now())

		query, args, err := inserMembers.ToSql()
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

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete chat: %w", err)
	}

	return nil
}

func (r *repo) SendMessage(ctx context.Context, msg MessageCreateDTO) error {
	// Добавим валидацию, что пользователь есть в чате, делаю запрос без использования squirrel
	var isMember bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM chat_members WHERE chat_id = $1 AND username = $2)", msg.ChatId, msg.FromUsername).Scan(&isMember)
	if err != nil {
		return fmt.Errorf("check chat membership: %w", err)
	}

	if !isMember {
		return errors.New("user is not a member of the chat")
	}

	// Выполняем основной запрос, добавляем в таблицу messages
	builder := squirrel.Insert("messages").
		PlaceholderFormat(squirrel.Dollar).
		Columns("chat_Id", "from_username", "text").
		Values(msg.ChatId, msg.FromUsername, msg.Text)
	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build send message query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	return nil
}

func (r *repo) RecentMessages(ctx context.Context, chatID int64, limit int) ([]MessageDTO, error) {
	builder := squirrel.Select("id", "chat_id", "from_username", "text", "created_at").
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
		err := rows.Scan(&message.Id, &message.ChatId, &message.From, &message.Text, &message.CreatedAt)
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

func (r *repo) UserChats(ctx context.Context, username string) ([]ChatInfoDTO, error) {
	builder := squirrel.Select("id", "created_at").
		PlaceholderFormat(squirrel.Dollar).
		From("chats").
		Join("chat_members ON chats.ID = chat_members.chat_id ").
		Where(squirrel.Eq{"username": username}).
		OrderBy("created_at DESC")

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
		err := rows.Scan(&chat.ID, &chat.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan chat: %w", err)
		}

		// Получаем участников чата
		members, err := r.chatMembers(ctx, chat.ID)
		if err != nil {
			return nil, fmt.Errorf("get members for chat %d: %w", chat.ID, err)
		}

		chat.Usernames = members
		chats = append(chats, chat)
	}

	return chats, nil
}

// Получаем участников чата
func (r *repo) chatMembers(ctx context.Context, chatID int64) ([]string, error) {
	builder := squirrel.Select("username").
		PlaceholderFormat(squirrel.Dollar).
		From("chat_members").
		Where(squirrel.Eq{"chat_id": chatID})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var usernames []string
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query chat members: %w", err)
	}

	for rows.Next() {
		var username string
		err := rows.Scan(&username)
		if err != nil {
			return nil, fmt.Errorf("scan username: %w", err)
		}
		usernames = append(usernames, username)
	}

	return usernames, nil
}
