package user

import (
	"auth/internal/model"
	userRepository "auth/internal/repository/user"
	"context"
)

type UserService interface {
	Create(ctx context.Context, info CreateUserDTO) (int64, error)
	Get(ctx context.Context, id int64) (*model.User, error)
	Update(ctx context.Context, id int64, info UpdateUserDTO) error
	Delete(ctx context.Context, id int64) error
}

type service struct {
	userRepository userRepository.UserRepository
}

func NewService(userRepository userRepository.UserRepository) UserService {
	return &service{userRepository: userRepository}
}
