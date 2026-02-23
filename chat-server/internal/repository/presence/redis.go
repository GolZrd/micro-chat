package presence

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	presencePrefix = "presence:user:"
	lastSeenPrefix = "lastseen:user:"
	onlineTTL      = 60 * time.Second // Если не обновили за 60 сек ставим автоматически offline
)

type UserPresence struct {
	UserId     int64
	IsOnline   bool
	LastSeenAt time.Time
}

type RedisRepository interface {
	SetOnline(ctx context.Context, userId int64) error
	GetPresence(ctx context.Context, userIds []int64) ([]UserPresence, error)
}

type RedisPresence struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) RedisRepository {
	return &RedisPresence{client: client}
}

// SetOnline ставим ключ с TTL, если не обновили за 60 сек, выставляем автоматически offline
func (r *RedisPresence) SetOnline(ctx context.Context, userId int64) error {
	pipe := r.client.Pipeline()

	// Собираем ключ и сохраняем с TTL
	presenceKey := fmt.Sprintf("%s%d", presencePrefix, userId)
	pipe.Set(ctx, presenceKey, "1", onlineTTL)

	// Обновляем lastseen
	lastSeenKey := fmt.Sprintf("%s%d", lastSeenPrefix, userId)
	pipe.Set(ctx, lastSeenKey, time.Now().Unix(), 30*24*time.Hour) // Будем хранить 30 дней

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("set online: %w", err)
	}

	return nil
}

// GetPresence батчевая проверка онлайн пользователей через pipeline
func (r *RedisPresence) GetPresence(ctx context.Context, userIds []int64) ([]UserPresence, error) {
	pipe := r.client.Pipeline()

	//Собираем команды для lastSeen и Presence
	presenceCmd := make(map[int64]*redis.IntCmd, len(userIds))
	lastSeenCmd := make(map[int64]*redis.StringCmd, len(userIds))

	for _, id := range userIds {
		presenceKey := fmt.Sprintf("%s%d", presencePrefix, id)
		presenceCmd[id] = pipe.Exists(ctx, presenceKey)

		lastSeenKey := fmt.Sprintf("%s%d", lastSeenPrefix, id)
		lastSeenCmd[id] = pipe.Get(ctx, lastSeenKey)
	}

	// Теперь делаем батчевый запрос к Redis
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("pipeline exec: %w", err)
	}

	res := make([]UserPresence, 0, len(userIds))

	for _, id := range userIds {
		user := UserPresence{
			UserId:   id,
			IsOnline: presenceCmd[id].Val() > 0,
		}

		if lastSeenStr, err := lastSeenCmd[id].Result(); err == nil && lastSeenStr != "" {
			if unix, err := strconv.ParseInt(lastSeenStr, 10, 64); err == nil {
				user.LastSeenAt = time.Unix(unix, 0)
			}
		}

		res = append(res, user)
	}

	return res, nil
}
