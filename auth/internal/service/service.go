package service

import (
	"auth/internal/model"
	"auth/internal/repository"
	"context"
)

type AuthService interface {
	Create(ctx context.Context, info CreateUserDTO) (int64, error)
	Get(ctx context.Context, id int64) (*model.User, error)
	Update(ctx context.Context, id int64, info UpdateUserDTO) error
	Delete(ctx context.Context, id int64) error
}

type service struct {
	authRepository repository.AuthRepository
}

func NewService(authRepository repository.AuthRepository) AuthService {
	return &service{authRepository: authRepository}
}
