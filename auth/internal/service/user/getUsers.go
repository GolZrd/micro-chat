package user

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/model"
	"go.uber.org/zap"
)

func (s *service) GetUsers(ctx context.Context, ids []int64) ([]model.UserShort, error) {
	users, err := s.userRepository.GetUsers(ctx, ids)
	if err != nil {
		logger.Error("Failed to get users", zap.Int64s("user_ids", ids), zap.Error(err))
		return nil, fmt.Errorf("failed to get users %w", err)
	}

	return users, nil
}
