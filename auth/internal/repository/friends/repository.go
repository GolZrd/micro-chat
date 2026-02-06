package friends

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/internal/repository/friends/model"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableName = "friends"

	idColumn        = "id"
	userIdColumn    = "user_id"
	friendIdColumn  = "friend_id"
	statusColumn    = "status"
	CreatedAtColumn = "created_at"
	UpdatedAtColumn = "updated_at"
)

type FriendsRepository interface {
	SendFriendRequest(ctx context.Context, userId, friendId int64) error          // Отправить запрос в друзья, передаем свой userId и id друга, которому отправляем запрос
	AcceptFriendRequest(ctx context.Context, requestId int64, userId int64) error // Передаются id запроса и id пользователя, которого добавляем в друзья
	RejectFriendRequest(ctx context.Context, requestId int64, userId int64) error
	Friends(ctx context.Context, userid int64) ([]model.Friend, error)
	RemoveFriend(ctx context.Context, userId, friendId int64) error
	FriendRequests(ctx context.Context, userId int64) ([]model.FriendRequest, error)
	FriendshipStatus(ctx context.Context, userId, otherUserId int64) (string, error) // Функция для проверки статуса дружбы
}

type repo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) FriendsRepository {
	return &repo{
		db: db,
	}
}

// Функция для проверки существующего статуса дружбы
func (r *repo) FriendshipStatus(ctx context.Context, userId, otherUserId int64) (string, error) {
	builder := squirrel.Select(statusColumn, userIdColumn).
		PlaceholderFormat(squirrel.Dollar).
		From(tableName).
		Where(squirrel.Or{
			squirrel.Eq{userIdColumn: userId, friendIdColumn: otherUserId},
			squirrel.Eq{userIdColumn: otherUserId, friendIdColumn: userId},
		}).
		Limit(1)

	query, args, err := builder.ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build query: %w", err)
	}

	var status string
	var senderId int64

	err = r.db.QueryRow(ctx, query, args...).Scan(&status, &senderId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "none", nil
		}
		return "", fmt.Errorf("failed to query row: %w", err)
	}

	if status == "accepted" {
		return "friends", nil
	}

	if senderId == userId {
		return "pending_sent", nil
	}

	return "pending_received", nil
}

func (r *repo) SendFriendRequest(ctx context.Context, userId, friendId int64) error {
	// Сначала проверяем статус дружбы, есть ли заяка в друзья
	status, _ := r.FriendshipStatus(ctx, userId, friendId)
	if status == "friends" {
		return fmt.Errorf("already friends")
	}
	if status == "pending_sent" {
		return fmt.Errorf("request already sent")
	}
	if status == "pending_received" {
		return fmt.Errorf("you have pending request from this user")
	}

	builder := squirrel.Insert(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Columns(userIdColumn, friendIdColumn, statusColumn).
		Values(userId, friendId, "pending")

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

func (r *repo) AcceptFriendRequest(ctx context.Context, requestId int64, userId int64) error {
	// Обновляем статус входящей заявки
	var fromUserId int64
	builderUpdate := squirrel.Update(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{idColumn: requestId, friendIdColumn: userId, statusColumn: "pending"}).
		SetMap(map[string]interface{}{
			statusColumn:    "accepted",
			UpdatedAtColumn: squirrel.Expr("NOW()"),
		}).Suffix("RETURNING user_id")

	query, args, err := builderUpdate.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	err = r.db.QueryRow(ctx, query, args...).Scan(&fromUserId)
	if err != nil {
		return fmt.Errorf("failed to query row: %w", err)
	}

	// Теперь создаем запись в бд для user_id, который принял заявку и у него появился новый друг и статус становится сразу accepted
	// То есть у нас была 1 запись, у того, кто отправил заявку, а после принятия заявки у отправившегося меняется только статус, а у того, кто принял появляется новая запись в БД со статусом дружбы

	builderInsert := squirrel.Insert(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Columns(userIdColumn, friendIdColumn, statusColumn).
		Values(userId, fromUserId, "accepted").
		Suffix("ON CONFLICT (user_id, friend_id) DO UPDATE SET status = 'accepted', updated_at = NOW()")

	query, args, err = builderInsert.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}

func (r *repo) RejectFriendRequest(ctx context.Context, requestId int64, userId int64) error {
	builder := squirrel.Delete(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{idColumn: requestId, friendIdColumn: userId, statusColumn: "pending"})

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("request not found")
	}

	return nil
}

func (r *repo) Friends(ctx context.Context, userid int64) ([]model.Friend, error) {
	// Запрос использующий Join удобнее просто написать без squirrel
	query := `SELECT f.ID, u.id, u.username
			FROM friends f 
			JOIN users u ON u.id = f.friend_id
			WHERE f.user_id = $1 AND f.status = 'accepted'
			ORDER BY u.username`

	rows, err := r.db.Query(ctx, query, userid)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var friends []model.Friend
	for rows.Next() {
		var friend model.Friend
		if err := rows.Scan(&friend.Id, &friend.UserId, &friend.Username); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		friends = append(friends, friend)
	}
	return friends, nil
}

// Удаляем из друзей
func (r *repo) RemoveFriend(ctx context.Context, userId, friendId int64) error {
	builder := squirrel.Delete(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Or{
			squirrel.Eq{userIdColumn: userId, friendIdColumn: friendId},
			squirrel.Eq{userIdColumn: friendId, friendIdColumn: userId},
		})

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

func (r *repo) FriendRequests(ctx context.Context, userId int64) ([]model.FriendRequest, error) {
	query := `SELECT f.ID, f.user_id, u.username, f.created_at
			FROM friends f 
			JOIN users u ON u.id = f.user_id
			WHERE f.friend_id = $1 AND f.status = 'pending'
			ORDER BY f.created_at DESC`

	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var requests []model.FriendRequest
	for rows.Next() {
		var req model.FriendRequest
		if err := rows.Scan(&req.Id, &req.FromUserId, &req.FromUsername, &req.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}
