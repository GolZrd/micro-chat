package user

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"go.uber.org/zap"
)

func (s *service) Delete(ctx context.Context, id int64) error {
	logger.Info(
		"Attempt to delete user",
		zap.Int64("user_id", id),
	)

	err := s.userRepository.Delete(ctx, id)
	if err != nil {
		logger.Error("Failed to delete user", zap.Int64("user_id", id), zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.Info(
		"User deleted successfully",
		zap.Int64("user_id", id),
	)

	return nil
}
