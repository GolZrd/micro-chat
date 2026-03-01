package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	repoChat "github.com/GolZrd/micro-chat/chat-server/internal/repository"
	"go.uber.org/zap"
)

type ErrUserNotFound struct {
	Usernames []string
}

func (e *ErrUserNotFound) Error() string {
	return fmt.Sprintf("users not found: %s", strings.Join(e.Usernames, ","))
}

func (s *service) Create(ctx context.Context, name string, isPublic bool, creatorId int64, creatorUsername string, usernames []string) (int64, error) {
	// Проверяем что Usernames не пусто
	if len(usernames) == 0 {
		return 0, errors.New("usernames cannot be empty")
	}

	// Убираем создателя из общего среза
	var otherUsernames []string
	for _, username := range usernames {
		if username != creatorUsername {
			otherUsernames = append(otherUsernames, username)
		}
	}

	// Добавляем создателя
	members := []repoChat.MemberDTO{
		{UserId: creatorId, Username: creatorUsername},
	}

	// Одним вызовом проверяем и получаем ID всех участников
	result, err := s.authClient.CheckUsersExists(ctx, otherUsernames)
	if err != nil {
		logger.Error("failed to check users exists", zap.Strings("usernames", otherUsernames), zap.Error(err))
		return 0, fmt.Errorf("failed to check users: %w", err)
	}

	// Если переданных пользователей не существуют, то возвращаем ошибку
	if len(result.NotFoundUsers) > 0 {
		logger.Warn("users not found", zap.Strings("usernames", result.NotFoundUsers))
		return 0, &ErrUserNotFound{Usernames: result.NotFoundUsers}
	}

	for _, u := range result.FoundUsers {
		members = append(members, repoChat.MemberDTO{
			UserId:   u.Id,
			Username: u.Username,
		})
	}

	// Проверяем поле name, если пустое, то генерируем имя
	if name == "" {
		name = generateChatName(otherUsernames)
	}

	// Проверяем что это групповой чат
	isGroup := len(members) > 2

	// Создаем DTO для создания чата
	dto := repoChat.CreateChatDTO{
		Name:      name,
		IsGroup:   isGroup,
		IsPublic:  isPublic,
		CreatorId: creatorId,
		Members:   members,
	}

	// Создаем чат с существующими участниками
	id, err := s.ChatRepository.Create(ctx, dto)
	if err != nil {
		logger.Error("failed to create chat",
			zap.Strings("usernames", otherUsernames),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to create chat: %w", err)
	}

	logger.Info("chat created successfully",
		zap.Int64("chat_id", id),
		zap.String("name", name),
		zap.Bool("is_Public", isPublic),
		zap.Int64("creator_id", creatorId),
		zap.Int("members_count", len(members)),
	)

	return id, nil
}

func generateChatName(usernames []string) string {
	if len(usernames) == 0 {
		return "Новый чат"
	}

	if len(usernames) <= 3 {
		return strings.Join(usernames, ", ")
	}

	return fmt.Sprintf("%s и еще %d", strings.Join(usernames[:2], ", "), len(usernames)-2)
}
