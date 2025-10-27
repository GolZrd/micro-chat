package user

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/model"
	"go.uber.org/zap"
)

func (s *service) Get(ctx context.Context, id int64) (*model.User, error) {

	user, err := s.userRepository.Get(ctx, id)
	if err != nil {
		logger.Error("Failed to get user", zap.Int64("user_id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get user %w", err)
	}

	logger.Info(
		"Get user",
		zap.Int64("user_id", id),
		zap.String("Name", user.Info.Name),
		zap.String("Role", user.Info.Role),
	)

	return user, nil
}
