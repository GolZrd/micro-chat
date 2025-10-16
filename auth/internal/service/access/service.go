package access

import (
	"context"

	"github.com/GolZrd/micro-chat/auth/internal/cache"
	"github.com/GolZrd/micro-chat/auth/internal/config"
	accessRepository "github.com/GolZrd/micro-chat/auth/internal/repository/access"
)

type AccessService interface {
	Check(ctx context.Context, accessToken string, endPoint string) error
}

type service struct {
	accessRepository accessRepository.AccessRepository
	redisCache       *cache.RedisCache
	AccessSecretKey  string
}

func NewService(accessRepository accessRepository.AccessRepository, redisCache *cache.RedisCache, cfg *config.Config) AccessService {
	return &service{accessRepository: accessRepository, redisCache: redisCache, AccessSecretKey: cfg.AccessSecretKey}
}
