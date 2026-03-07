package unread

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UnreadRepository interface {
	IncrementUnreadForMembers(ctx context.Context, chatId, senderId int64) error
	MarkAsRead(ctx context.Context, chatId, userId int64) error
	AllUnreadCounts(ctx context.Context, userId int64) (map[int64]int32, error)
	DeleteForChat(ctx context.Context, chatId int64) error
	InitForMember(ctx context.Context, chatId, userId int64) error
}

type unreadRepo struct {
	db *pgxpool.Pool
}

func NewUnreadRepository(db *pgxpool.Pool) UnreadRepository {
	return &unreadRepo{db: db}
}

// IncrementUnreadForMembers - увеличить счётчик для всех участников кроме отправителя.
func (r *unreadRepo) IncrementUnreadForMembers(ctx context.Context, chatId, senderId int64) error {
	query := `
		INSERT INTO chat_unread (chat_id, user_id, count)
		SELECT $1, cm.user_id, 1
		FROM chat_members cm
		WHERE cm.chat_id = $1 AND cm.user_id != $2
		ON CONFLICT (chat_id, user_id)
		DO UPDATE SET count = chat_unread.count + 1
	`
	_, err := r.db.Exec(ctx, query, chatId, senderId)
	if err != nil {
		return fmt.Errorf("increment unread for members: %w", err)
	}
	return nil
}

// MarkAsRead - обнулить счётчик непрочитанных
func (r *unreadRepo) MarkAsRead(ctx context.Context, chatId, userId int64) error {
	query := `
		INSERT INTO chat_unread (chat_id, user_id, count, last_read_at)
		VALUES ($1, $2, 0, NOW())
		ON CONFLICT (chat_id, user_id)
		DO UPDATE SET count = 0, last_read_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, chatId, userId)
	if err != nil {
		return fmt.Errorf("mark as read: %w", err)
	}
	return nil
}

// AllUnreadCounts - получаем все непрочитанные для пользователя
func (r *unreadRepo) AllUnreadCounts(ctx context.Context, userId int64) (map[int64]int32, error) {
	builder := squirrel.Select("chat_id", "count").
		PlaceholderFormat(squirrel.Dollar).
		From("chat_unread").
		Where(squirrel.And{
			squirrel.Eq{"user_id": userId},
			squirrel.Gt{"count": 0},
		})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query unread counts: %w", err)
	}
	defer rows.Close()

	counts := make(map[int64]int32)
	for rows.Next() {
		var chatId int64
		var count int32
		err := rows.Scan(&chatId, &count)
		if err != nil {
			return nil, fmt.Errorf("scan unread count: %w", err)
		}
		counts[chatId] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return counts, nil
}

// DeleteForChat - удаляем все записи непрочитанных при удалении чата
func (r *unreadRepo) DeleteForChat(ctx context.Context, chatId int64) error {
	builder := squirrel.Delete("chat_unread").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"chat_id": chatId})

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build delete query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete unread for chat: %w", err)
	}

	return nil
}

// InitForMember - создаем запись при добавлении участника в чат
func (r *unreadRepo) InitForMember(ctx context.Context, chatId, userId int64) error {
	query := `
		INSERT INTO chat_unread (chat_id, user_id, count)
		VALUES ($1, $2, 0)
		ON CONFLICT (chat_id, user_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, chatId, userId)
	if err != nil {
		return fmt.Errorf("init unread for member: %w", err)
	}
	return nil
}
