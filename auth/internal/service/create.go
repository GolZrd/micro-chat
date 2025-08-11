package service

import (
	"auth/internal/repository"
	"context"
	"errors"
)

func (s *service) Create(ctx context.Context, info CreateUserDTO) (int64, error) {
	if info.Password != info.PasswordConfirm {
		return 0, errors.New("passwords do not match")
	}

	// service DTO → repository DTO
	params := repository.CreateUserDTO{
		Name:     info.Name,
		Email:    info.Email,
		Password: info.Password,
		Role:     info.Role,
	}

	id, err := s.authRepository.Create(ctx, params)
	if err != nil {
		return 0, err
	}

	return id, nil
}
