package service

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

func (s *service) Delete(ctx context.Context, id int64) error {
	err := s.ChatRepository.Delete(ctx, id)
	if err != nil {
		logger.Error("failed to delete chat",
			zap.Int64("chat_id", id),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	logger.Info(
		"chat deleted successfully",
		zap.Int64("chat_id", id),
	)

	return nil
}
