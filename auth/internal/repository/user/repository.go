package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/model"
	"github.com/GolZrd/micro-chat/auth/internal/repository/user/converter"
	modelRepo "github.com/GolZrd/micro-chat/auth/internal/repository/user/model"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableName = "users"

	idColumn        = "id"
	usernameColumn  = "username"
	emailColumn     = "email"
	passwordColumn  = "password"
	roleColumn      = "role"
	avatarColumn    = "avatar_url"
	bioColumn       = "bio"
	CreatedAtColumn = "created_at"
	UpdatedAtColumn = "updated_at"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(ctx context.Context, info CreateUserDTO) (int64, error)
	Get(ctx context.Context, id int64) (*model.User, error)
	Update(ctx context.Context, id int64, info UpdateUserDTO) error
	Delete(ctx context.Context, id int64) error
	GetByEmail(ctx context.Context, email string) (*model.UserAuthData, error)
	GetByUsernames(ctx context.Context, usernames []string) ([]model.UserShort, error)
	GetUsers(ctx context.Context, ids []int64) ([]model.UserShort, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	SearchUser(ctx context.Context, searchQuery string, currentUserId int64, limit int) ([]model.UserSearchResult, error)
	UpdateAvatar(ctx context.Context, id int64, avatarUrl string) error
}

type repo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) UserRepository {
	return &repo{
		db: db,
	}
}

// Create - метод для создания нового пользователя в БД
func (r *repo) Create(ctx context.Context, info CreateUserDTO) (int64, error) {
	builder := squirrel.Insert(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Columns(usernameColumn, emailColumn, passwordColumn, roleColumn).
		Values(info.Username, info.Email, info.Password, info.Role).
		Suffix("RETURNING id")

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	var id int64
	err = r.db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to query row: %w", err)
	}

	return id, nil
}

func (r *repo) Get(ctx context.Context, id int64) (*model.User, error) {
	builder := squirrel.Select(idColumn, usernameColumn, passwordColumn, emailColumn, roleColumn, avatarColumn, bioColumn, CreatedAtColumn, UpdatedAtColumn).
		PlaceholderFormat(squirrel.Dollar).
		From(tableName).
		Where(squirrel.Eq{idColumn: id}).
		Limit(1)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var user modelRepo.User
	err = r.db.QueryRow(ctx, query, args...).Scan(&user.Id, &user.Info.Username, &user.Info.Password, &user.Info.Email, &user.Info.Role, &user.AvatarURL, &user.Bio, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to query row: %w", err)
	}

	return converter.ToUserFromRepo(&user), nil
}

func (r *repo) Update(ctx context.Context, id int64, info UpdateUserDTO) error {
	builder := squirrel.Update(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{idColumn: id}).
		Set(UpdatedAtColumn, time.Now())

	hasUpdates := false

	// Обновляем ТОЛЬКО переданные поля
	if info.Username != nil {
		builder = builder.Set(usernameColumn, *info.Username)
		hasUpdates = true
	}
	if info.Email != nil {
		builder = builder.Set(emailColumn, *info.Email)
		hasUpdates = true
	}
	if info.Bio != nil {
		builder = builder.Set(bioColumn, *info.Bio)
		hasUpdates = true
	}

	if !hasUpdates {
		return nil
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
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
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (r *repo) GetByEmail(ctx context.Context, email string) (*model.UserAuthData, error) {
	builder := squirrel.Select(idColumn, usernameColumn, passwordColumn, emailColumn, roleColumn).
		PlaceholderFormat(squirrel.Dollar).
		From(tableName).
		Where(squirrel.Eq{emailColumn: email}).
		Limit(1)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var userData modelRepo.UserAuthData
	err = r.db.QueryRow(ctx, query, args...).Scan(&userData.Id, &userData.Username, &userData.Password, &userData.Email, &userData.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query row: %w", err)
	}

	return converter.ToUserAuthDataFromRepo(&userData), nil
}

// GetByUsernames возвращает имена существующих пользователей
func (r *repo) GetByUsernames(ctx context.Context, usernames []string) ([]model.UserShort, error) {
	// Простой запрос, потому что проверяем только имя, squirrel builder избыточен
	query := "SELECT id, username, COALESCE(avatar_url, '') FROM users WHERE username = ANY($1)"

	rows, err := r.db.Query(ctx, query, usernames)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var existingUsers []model.UserShort
	for rows.Next() {
		var user modelRepo.UserShort
		if err := rows.Scan(&user.Id, &user.Username, &user.AvatarUrl); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		existingUsers = append(existingUsers, *converter.ToUserShortFromRepo(&user))
	}
	// Возвращаются только те пользователи, которые есть в БД
	return existingUsers, nil
}

func (r *repo) GetUsers(ctx context.Context, ids []int64) ([]model.UserShort, error) {
	builder := squirrel.Select(idColumn, usernameColumn, "COALESCE(avatar_url, '')").
		PlaceholderFormat(squirrel.Dollar).
		From(tableName).
		Where(squirrel.Eq{idColumn: ids})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var users []model.UserShort
	for rows.Next() {
		var user modelRepo.UserShort
		if err := rows.Scan(&user.Id, &user.Username, &user.AvatarUrl); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		users = append(users, *converter.ToUserShortFromRepo(&user))
	}

	return users, nil
}

func (r *repo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := "SELECT id, username, email, created_at FROM users WHERE username = $1"

	var user modelRepo.User
	err := r.db.QueryRow(ctx, query, username).Scan(&user.Id, &user.Info.Username, &user.Info.Email, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query row: %w", err)
	}

	return converter.ToUserFromRepo(&user), nil
}

// SearchUser метод для поиска пользователей, передаем query и id текущего пользователя, чтобы не показывать его в результате
// Запрос будет строиться с JOIN, будем присоединять таблицу с друзьями к таблице с пользователями
func (r *repo) SearchUser(ctx context.Context, searchQuery string, currentUserId int64, limit int) ([]model.UserSearchResult, error) {
	// Проверяем лимиты
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	// Собираем паттерн для поиска
	searchPattern := "%" + searchQuery + "%"

	// Строим запрос с JOIN, будем присоединять таблицу с друзьями к таблице с пользователями
	builder := squirrel.Select(
		"u.id",
		"u.username",
		"COALESCE(f.status, 'none') as friendship_status",
		"COALESCE(u.avatar_url, '') as avatar_url"). // COALESCE - возвращает первое не NULL значение, то есть, если связи нет, вернет "none"
		PlaceholderFormat(squirrel.Dollar).
		From("users u").
		LeftJoin("friends f ON f.friend_id = u.id AND f.user_id = ?", currentUserId). // f.user_id - текущий пользователь, f.friend_id - пользователь, которого мы ищем
		Where(squirrel.And{
			squirrel.NotEq{"u.id": currentUserId}, // Исключаем текущего пользователя
			squirrel.ILike{"u.username": searchPattern},
		}).
		OrderBy("u.username").
		Limit(uint64(limit))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	users := make([]model.UserSearchResult, 0)

	for rows.Next() {
		var user model.UserSearchResult
		if err := rows.Scan(&user.Id, &user.Username, &user.FriendshipStatus, &user.AvatarURL); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateAvatar обновляет аватар пользователя
func (r *repo) UpdateAvatar(ctx context.Context, id int64, avatarUrl string) error {
	query := "UPDATE users SET avatar_url = $1, updated_at = now() WHERE id = $2"
	_, err := r.db.Exec(ctx, query, avatarUrl, id)
	if err != nil {
		return fmt.Errorf("failed to update avatar: %w", err)
	}
	return nil
}
