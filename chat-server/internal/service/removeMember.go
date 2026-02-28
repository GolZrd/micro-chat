package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

func (s *service) RemoveMember(ctx context.Context, chatId int64, userId int64, targetUserId int64) error {
	// Получаем информацию о чате
	chat, err := s.ChatRepository.ChatInfo(ctx, chatId)
	if err != nil {
		logger.Error("failed to get chat info", zap.Int64("chat_id", chatId), zap.Error(err))
		return fmt.Errorf("get chat info: %w", err)
	}

	if chat.IsDirect {
		logger.Warn("cannot remove members from direct chat", zap.Int64("chat_id", chatId))
		return errors.New("cannot remove members from direct chat")
	}

	if chat.CreatorId != userId {
		logger.Warn("only owner can remove members from chat", zap.Int64("chat_id", chatId))
		return errors.New("only owner can remove members from chat")
	}

	// Нельзя удалить самого себя
	if userId == targetUserId {
		logger.Warn("cannot remove yourself from chat", zap.Int64("chat_id", chatId))
		return errors.New("cannot remove yourself from chat")
	}

	err = s.ChatRepository.RemoveMember(ctx, chatId, targetUserId)
	if err != nil {
		logger.Error("failed to remove member", zap.Int64("chat_id", chatId), zap.Int64("user_id", targetUserId), zap.Error(err))
		return fmt.Errorf("remove member: %w", err)
	}

	logger.Info("member removed from chat", zap.Int64("chat_id", chatId), zap.Int64("user_id", targetUserId))

	return nil
}
