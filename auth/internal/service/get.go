package service

import (
	"auth/internal/model"
	"context"
)

func (s *service) Get(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.authRepository.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}
