package user

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"go.uber.org/zap"
)

func (s *service) UpdateAvatar(ctx context.Context, id int64, avatarUrl string) error {
	err := s.userRepository.UpdateAvatar(ctx, id, avatarUrl)
	if err != nil {
		logger.Error("failed to update avatar", zap.Int64("user_id", id), zap.Error(err))
		return fmt.Errorf("failed to update avatar: %w", err)
	}
	return nil
}
