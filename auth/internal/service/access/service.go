package access

import (
	"context"

	"github.com/GolZrd/micro-chat/auth/internal/cache"
	"github.com/GolZrd/micro-chat/auth/internal/config"
	accessRepository "github.com/GolZrd/micro-chat/auth/internal/repository/access"
	"github.com/GolZrd/micro-chat/auth/internal/utils/jwt"
)

type AccessService interface {
	Check(ctx context.Context, accessToken string, endPoint string) error
}

type service struct {
	accessRepository accessRepository.AccessRepository
	jwtManager       jwt.JWTManager
	redisCache       *cache.RedisCache
	AccessSecretKey  string
}

func NewService(accessRepository accessRepository.AccessRepository, jwtManager jwt.JWTManager, redisCache *cache.RedisCache, cfg *config.Config) AccessService {
	return &service{accessRepository: accessRepository, jwtManager: jwtManager, redisCache: redisCache, AccessSecretKey: cfg.AccessSecretKey}
}
