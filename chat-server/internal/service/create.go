package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

func (s *service) Create(ctx context.Context, usernames []string) (int64, error) {
	// Проверяем что Usernames не пусто
	if len(usernames) == 0 {
		return 0, errors.New("usernames cannot be empty")
	}

	// Пока будем просто будем создавать чат с переданными usernames, дальше нужно будет изменить
	// TODO: первым делом нужно будет проверить что пользователи существуют, с помощью запроса к сервису AUTH, где у нас регистрируются и создаются пользователи

	id, err := s.ChatRepository.Create(ctx, usernames)
	if err != nil {
		logger.Error("failed to create chat",
			zap.Strings("usernames", usernames),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to create chat: %w", err)
	}

	logger.Info(
		"chat created successfully",
		zap.Strings("usernames", usernames),
	)

	return id, nil
}
