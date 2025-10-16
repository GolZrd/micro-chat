package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

// Инициализируем RedisCache
func NewRedisCache(addr string, password string, db int, ttl time.Duration) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Проверяем подключение
	err := client.Ping(context.Background()).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ttl:    ttl,
	}, nil

}

// Определим методы получения и сохранения данных в RedisCache
// GetEndpointRoles получает мапу ролей для конкретного эндпоинта
func (r *RedisCache) GetEndpointRoles(ctx context.Context, endpoint string) (map[string]struct{}, error) {
	// Пытаемся получить данные из Redis
	data, err := r.client.Get(ctx, endpoint).Result()
	if err == redis.Nil {
		// Ключ не найден
		log.Printf("Cache miss for endoint : %s", endpoint)
		return nil, nil
	}
	if err != nil {
		log.Printf("Redis error for %s: %v", endpoint, err)
		return nil, err
	}

	var roles []string

	// Пытаемся декодировать в срез roles
	if err := json.Unmarshal([]byte(data), &roles); err != nil {
		log.Printf("failed to unmarshal cached data for %s: %v", endpoint, err)
		return nil, err
	}

	// Теперь преобразуем в мапу ролей
	rolesMap := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		rolesMap[role] = struct{}{}
	}

	log.Printf("Cache HIT for endpoint %s: roles=%v", endpoint, roles)

	return rolesMap, nil
}

// SetEndpointRoles сохраняет мапу ролей для конкретного эндпоинта в RedisCache
func (r *RedisCache) SetEndpointRoles(ctx context.Context, endpoint string, roles map[string]struct{}) error {

	// преобразуем мапу в срез ролей
	rolesSlice := make([]string, 0, len(roles))
	for role := range roles {
		rolesSlice = append(rolesSlice, role)
	}

	// Переводим в JSON
	data, err := json.Marshal(rolesSlice)
	if err != nil {
		return fmt.Errorf("failed to marshal roles: %w", err)
	}

	// Сохраняем в Redis с утановленным TTL
	if err := r.client.Set(ctx, endpoint, data, r.ttl).Err(); err != nil {
		log.Printf("failed to cache enpoint: %s with roles: %v", endpoint, err)
		return err
	}

	log.Printf("Cached endpoint %s with roles: %v and TTL: %s", endpoint, rolesSlice, r.ttl)
	return nil
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}
