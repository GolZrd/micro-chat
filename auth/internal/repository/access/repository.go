package access

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableName      = "role_permissions"
	roleColumn     = "role"
	endPointColumn = "endpoint"
)

type AccessRepository interface {
	EndPointRoles(ctx context.Context, endPoint string) (map[string]struct{}, error)
}

type repo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) AccessRepository {
	return &repo{
		db: db,
	}
}

// Описываем метод
func (r *repo) EndPointRoles(ctx context.Context, endPoint string) (map[string]struct{}, error) {
	builder := squirrel.Select(roleColumn).
		PlaceholderFormat(squirrel.Dollar).
		From(tableName).
		Where(squirrel.Eq{endPointColumn: endPoint})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	roles := make(map[string]struct{})
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}

		roles[role] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}
