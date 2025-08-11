package service

import (
	"auth/internal/repository"
	"context"
)

func (s *service) Update(ctx context.Context, id int64, info UpdateUserDTO) error {
	// service DTO → repository DTO
	params := repository.UpdateUserDTO{
		Name:  info.Name,
		Email: info.Email,
	}

	err := s.authRepository.Update(ctx, id, params)
	if err != nil {
		return err
	}

	return nil
}
