package service

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

func (s *service) Heartbeat(ctx context.Context, userId int64) error {

	err := s.PresenceRepository.SetOnline(ctx, userId)
	if err != nil {
		logger.Error("Failed to set online", zap.Int64("user_id", userId), zap.Error(err))
		return fmt.Errorf("failed to set online: %w", err)
	}

	return nil
}
