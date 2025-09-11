package user

import (
	"context"

	userRepository "github.com/GolZrd/micro-chat/auth/internal/repository/user"
)

func (s *service) Update(ctx context.Context, id int64, info UpdateUserDTO) error {
	// service DTO → repository DTO
	params := userRepository.UpdateUserDTO{
		Name:  info.Name,
		Email: info.Email,
	}

	err := s.userRepository.Update(ctx, id, params)
	if err != nil {
		return err
	}

	return nil
}
