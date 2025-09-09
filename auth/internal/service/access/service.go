package access

import (
	"auth/internal/config"
	accessRepository "auth/internal/repository/access"
	"context"
)

type AccessService interface {
	Check(ctx context.Context, accessToken string, endPoint string) error
}

type service struct {
	accessRepository accessRepository.AccessRepository
	AccessSecretKey  string
}

func NewService(accessRepository accessRepository.AccessRepository, cfg *config.Config) AccessService {
	return &service{accessRepository: accessRepository, AccessSecretKey: cfg.AccessSecretKey}
}
