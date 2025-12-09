package user

import (
	"context"

	"github.com/GolZrd/micro-chat/auth/internal/model"
	userRepository "github.com/GolZrd/micro-chat/auth/internal/repository/user"
)

type UserService interface {
	Create(ctx context.Context, info CreateUserDTO) (int64, error)
	Get(ctx context.Context, id int64) (*model.User, error)
	Update(ctx context.Context, id int64, info UpdateUserDTO) error
	Delete(ctx context.Context, id int64) error
	CheckUsersExists(ctx context.Context, usernames []string) (notFoundUsers []string, err error)
}

type service struct {
	userRepository userRepository.UserRepository
}

func NewService(userRepository userRepository.UserRepository) UserService {
	return &service{userRepository: userRepository}
}
