package access

import (
	"context"
	"fmt"

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
		return nil, fmt.Errorf("build query for endpoint %s: %w", endPoint, err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query endpoint roles for %s: %w", endPoint, err)
	}

	roles := make(map[string]struct{})
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		roles[role] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return roles, nil
}
