package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

func (s *service) JoinChat(ctx context.Context, chatId int64, userId int64, username string) error {
	// Получаем информацию о чате
	chat, err := s.ChatRepository.ChatInfo(ctx, chatId)
	if err != nil {
		logger.Error("failed to get chat info", zap.Int64("chat_id", chatId), zap.Error(err))
		return fmt.Errorf("get chat info: %w", err)
	}

	if !chat.IsPublic {
		logger.Warn("cannot join private chat", zap.Int64("chat_id", chatId))
		return errors.New("cannot join private chat")
	}

	if chat.IsDirect {
		logger.Warn("cannot join direct chat", zap.Int64("chat_id", chatId))
		return errors.New("cannot join direct chat")
	}

	// Проверяем не участник ли уже в чате
	for _, m := range chat.Members {
		if m.UserId == userId {
			logger.Warn("user already in chat", zap.Int64("user_id", userId), zap.Int64("chat_id", chatId))
			return errors.New("user already in chat")
		}
	}

	err = s.ChatRepository.AddMember(ctx, chatId, userId, username)
	if err != nil {
		logger.Error("failed to add member", zap.Int64("chat_id", chatId), zap.Int64("user_id", userId), zap.Error(err))
		return fmt.Errorf("add member: %w", err)
	}

	logger.Info("User joined chat", zap.Int64("user_id", userId), zap.Int64("chat_id", chatId))

	return nil
}
