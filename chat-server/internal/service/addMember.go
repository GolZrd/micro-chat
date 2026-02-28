package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

func (s *service) AddMember(ctx context.Context, chatId int64, userId int64, username string) error {
	// Получаем информацию о чате
	chat, err := s.ChatRepository.ChatInfo(ctx, chatId)
	if err != nil {
		logger.Error("failed to get chat info", zap.Int64("chat_id", chatId), zap.Error(err))
		return fmt.Errorf("get chat info: %w", err)
	}

	// В личный чат добавлять нельзя
	if chat.IsDirect {
		logger.Warn("cannot add members to direct chat", zap.Int64("chat_id", chatId))
		return errors.New("cannot add members to direct chat")
	}

	// Если закрытый чат, то только владелец может добавить
	if !chat.IsPublic && chat.CreatorId != userId {
		logger.Warn("only owner can add members to private chat", zap.Int64("chat_id", chatId))
		return errors.New("only owner can add members to private chat")
	}

	// Проверяем что пользователь с таким именем существует
	res, err := s.authClient.CheckUsersExists(ctx, []string{username})
	if err != nil {
		logger.Error("failed to check user exists", zap.Strings("username", []string{username}), zap.Error(err))
		return fmt.Errorf("check users exists: %w", err)
	}

	if len(res.NotFoundUsers) > 0 {
		logger.Warn("users not found", zap.Strings("usernames", res.NotFoundUsers))
		return &ErrUserNotFound{Usernames: res.NotFoundUsers}
	}

	// Доп проверка, чтобы дальше получить ID
	if len(res.FoundUsers) == 0 {
		logger.Warn("user not found", zap.String("username", username))
		return &ErrUserNotFound{Usernames: []string{username}}
	}
	user := res.FoundUsers[0]

	// Добавляем пользователя в чат
	err = s.ChatRepository.AddMember(ctx, chatId, user.Id, user.Username)
	if err != nil {
		logger.Error("failed to add member to chat", zap.Int64("chat_id", chatId), zap.Int64("user_id", user.Id), zap.Error(err))
		return fmt.Errorf("add member to chat: %w", err)
	}

	return nil

}
