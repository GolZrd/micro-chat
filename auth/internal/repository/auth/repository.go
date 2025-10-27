package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableName       = "refresh_tokens"
	userIdColumn    = "user_id"
	tokenColumn     = "token"
	revokedColumn   = "revoked"
	createdAtColumn = "created_at"
	expiresAtColumn = "expires_at"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type AuthRepository interface {
	RevokeAllByUserId(ctx context.Context, userId int64) error                                      // Метод для обнуления всех токенов пользователя
	RevokeToken(ctx context.Context, token string) error                                            // Метод для обнуления определенного токена
	CreateRefreshToken(ctx context.Context, userId int64, token string, expiers_at time.Time) error // Метод для сохранения нового токена, используется для refreshToken
}

type repo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) AuthRepository {
	return &repo{
		db: db,
	}
}

// Описываем методы
// Метод для обнуления всех токенов пользователя
func (r *repo) RevokeAllByUserId(ctx context.Context, userId int64) error {
	builder := squirrel.Update(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{userIdColumn: userId}).
		Where(squirrel.Eq{revokedColumn: false}).
		Set(revokedColumn, true)

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build RevokeAllByUserId: %w", err)
	}

	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to exec RevokeAllByUserId userId=%d: %w", userId, err)
	}

	return nil
}

// Метод для обнуления токена по самому токену
func (r *repo) RevokeToken(ctx context.Context, token string) error {
	builder := squirrel.Update(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{tokenColumn: token}).
		Set(revokedColumn, true)

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build RevokeToken: %w", err)
	}

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to exec RevokeToken: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("refresh token not found")
	}

	return nil
}

// Метод для сохранения нового токена
func (r *repo) CreateRefreshToken(ctx context.Context, userId int64, token string, expiers_at time.Time) error {
	builder := squirrel.Insert(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Columns(userIdColumn, tokenColumn, expiresAtColumn).
		Values(userId, token, expiers_at)

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build CreateRefreshToken: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to exec CreateRefreshToken: %w", err)
	}

	return nil
}
