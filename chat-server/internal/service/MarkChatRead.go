package service

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

func (s *service) MarkChatRead(ctx context.Context, chatId, userId int64) error {
	err := s.UnreadRepository.MarkAsRead(ctx, chatId, userId)
	if err != nil {
		logger.Error("failed to mark chat as read", zap.Int64("chat_id", chatId), zap.Int64("user_id", userId), zap.Error(err))
		return fmt.Errorf("mark chat as read: %w", err)
	}
	return nil
}
