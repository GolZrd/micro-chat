package repository

import (
	"context"
	"errors"
	"time"

	"auth/internal/model"
	"auth/internal/repository/converter"
	modelRepo "auth/internal/repository/model"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableName = "users"

	idColumn        = "id"
	nameColumn      = "name"
	emailColumn     = "email"
	passwordColumn  = "password"
	roleColumn      = "role"
	CreatedAtColumn = "created_at"
	UpdatedAtColumn = "updated_at"
)

type AuthRepository interface {
	Create(ctx context.Context, info CreateUserDTO) (int64, error)
	Get(ctx context.Context, id int64) (*model.User, error)
	Update(ctx context.Context, id int64, info UpdateUserDTO) error
	Delete(ctx context.Context, id int64) error
}

type repo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) AuthRepository {
	return &repo{
		db: db,
	}
}

// Create - метод для создания нового пользователя в БД
func (r *repo) Create(ctx context.Context, info CreateUserDTO) (int64, error) {
	builder := squirrel.Insert(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Columns(nameColumn, emailColumn, passwordColumn, roleColumn).
		Values(info.Name, info.Email, info.Password, info.Role).
		Suffix("RETURNING id")

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	var id int64
	err = r.db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repo) Get(ctx context.Context, id int64) (*model.User, error) {
	builder := squirrel.Select(idColumn, nameColumn, passwordColumn, emailColumn, roleColumn, CreatedAtColumn, UpdatedAtColumn).
		PlaceholderFormat(squirrel.Dollar).
		From(tableName).
		Where(squirrel.Eq{idColumn: id}).
		Limit(1)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var user modelRepo.User
	err = r.db.QueryRow(ctx, query, args...).Scan(&user.Id, &user.Info.Name, &user.Info.Password, &user.Info.Email, &user.Info.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return converter.ToUserFromRepo(&user), nil
}

func (r *repo) Update(ctx context.Context, id int64, info UpdateUserDTO) error {
	builder := squirrel.Update(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{idColumn: id}).
		SetMap(map[string]interface{}{
			nameColumn:      info.Name,
			emailColumn:     info.Email,
			UpdatedAtColumn: time.Now(),
		})

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	// Проверяем обновились ли данные
	if res.RowsAffected() == 0 {
		return errors.New("user not found or no changes")
	}

	return nil
}

func (r *repo) Delete(ctx context.Context, id int64) error {
	builder := squirrel.Delete(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{idColumn: id})

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
