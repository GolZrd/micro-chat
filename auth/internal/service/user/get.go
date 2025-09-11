package user

import (
	"context"

	"github.com/GolZrd/micro-chat/auth/internal/model"
)

func (s *service) Get(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.userRepository.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}
