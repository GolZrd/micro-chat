package service

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

func (s *service) GetOrCreateDirectChat(ctx context.Context, currentUserId int64, currentUsername string, targetUserId int64, targetUsername string) (int64, bool, error) {
	// Проверям что не собираемся создать чат с самим собой
	if currentUserId == targetUserId {
		logger.Warn("Cannot create chat with yourself", zap.Int64("CurrentUserId", currentUserId), zap.Int64("TargetUserId", targetUserId))
		return 0, false, fmt.Errorf("cannot create chat with yourself")
	}

	// Пробуем найти существующий чат
	chatId, err := s.ChatRepository.FindDirectChat(ctx, currentUserId, targetUserId)
	if err != nil {
		logger.Error("Failed to find direct chat", zap.Int64("CurrentUserId", currentUserId), zap.Int64("TargetUserId", targetUserId), zap.Error(err))
		return 0, false, fmt.Errorf("failed to find direct chat: %w", err)
	}

	// Если нашли, то возвращаем его
	if chatId != 0 {
		logger.Debug("direct chat found", zap.Int64("chat_id", chatId), zap.Int64("user1", currentUserId), zap.Int64("user2", targetUserId))
		return chatId, false, nil
	}

	// Если не нашли, то создаем
	chatId, err = s.ChatRepository.CreateDirectChat(ctx, currentUserId, targetUserId, currentUsername, targetUsername)
	if err != nil {
		logger.Error("Failed to create direct chat", zap.Int64("CurrentUserId", currentUserId), zap.Int64("TargetUserId", targetUserId), zap.Error(err))
		return 0, false, fmt.Errorf("failed to create direct chat: %w", err)
	}

	logger.Info("direct chat created", zap.Int64("chat_id", chatId), zap.Int64("user1", currentUserId), zap.Int64("user2", targetUserId))
	return chatId, true, nil
}
