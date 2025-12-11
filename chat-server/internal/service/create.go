package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

type ErrUserNotFound struct {
	Usernames []string
}

func (e *ErrUserNotFound) Error() string {
	return fmt.Sprintf("users not found: %s", strings.Join(e.Usernames, ","))
}

func (s *service) Create(ctx context.Context, creatorUsername string, usernames []string) (int64, error) {
	// Проверяем что Usernames не пусто
	if len(usernames) == 0 {
		return 0, errors.New("usernames cannot be empty")
	}

	// Проверяем что переданные пользователи существуют
	notFoundUsers, err := s.authClient.CheckUsersExists(ctx, usernames)
	if err != nil {
		logger.Error("failed to check users exists", zap.Strings("usernames", usernames), zap.Error(err))
		return 0, fmt.Errorf("failed to check users: %w", err)
	}

	// Если переданных пользователей не существуют, то возвращаем ошибку
	if len(notFoundUsers) > 0 {
		logger.Warn("users not found", zap.Strings("usernames", notFoundUsers))
		return 0, &ErrUserNotFound{Usernames: notFoundUsers}
	}

	// Собираем всех участников
	participants := make([]string, 0, len(usernames)+1)
	participants = append(participants, creatorUsername)
	participants = append(participants, usernames...)

	// Создаем чат с существующими участниками
	id, err := s.ChatRepository.Create(ctx, participants)
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
