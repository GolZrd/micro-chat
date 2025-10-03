package user

import (
	"context"
	"errors"
	"strings"

	userRepository "github.com/GolZrd/micro-chat/auth/internal/repository/user"
)

func (s *service) Create(ctx context.Context, info CreateUserDTO) (int64, error) {
	if info.Password != info.PasswordConfirm {
		return 0, errors.New("passwords do not match")
	}

	// Нужно привести email к нижнему регистру
	email := strings.ToLower(info.Email)

	// service DTO → repository DTO
	params := userRepository.CreateUserDTO{
		Name:     info.Name,
		Email:    email,
		Password: info.Password,
		Role:     info.Role,
	}

	id, err := s.userRepository.Create(ctx, params)
	if err != nil {
		return 0, err
	}

	return id, nil
}
