package user

import (
	userRepository "auth/internal/repository/user"
	"context"
)

func (s *service) Update(ctx context.Context, id int64, info UpdateUserDTO) error {
	// service DTO → repository DTO
	params := userRepository.UpdateUserDTO{
		Name:  info.Name,
		Email: info.Email,
	}

	err := s.authRepository.Update(ctx, id, params)
	if err != nil {
		return err
	}

	return nil
}
