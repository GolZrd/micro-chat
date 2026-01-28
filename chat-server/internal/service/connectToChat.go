package service

import (
	"context"
	"fmt"
	"time"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

// ConnectToChat подписывает на получение сообщений из чата
func (s *service) ConnectToChat(ctx context.Context, userId int64, username string, chatId int64) (<-chan MessageDTO, error) {

	logger.Info("Connecting to chat", zap.Int64("user_id", userId), zap.String("username", username), zap.Int64("chat_id", chatId))

	// Проверяем что чат существует
	exists, err := s.ChatRepository.ChatExists(ctx, chatId)
	if err != nil {
		logger.Error("Failed to check chat exists", zap.Int64("chat_id", chatId), zap.Error(err))
		return nil, fmt.Errorf("check chat exists: %w", err)
	}
	if !exists {
		logger.Warn("chat not found", zap.Int64("chat_id", chatId))
		return nil, fmt.Errorf("chat %d not found", chatId)
	}

	// Получаем или создаем комнату
	room := s.getOrCreateRoom(chatId)

	// Создаем самого подписчика
	msgChan := make(chan MessageDTO, 100)
	sub := &Subscriber{
		Channel:  msgChan,
		UserId:   userId,
		Username: username,
		JoinedAt: time.Now(),
	}

	oldChannel := room.AddSubscriber(sub)

	// Закрываем старое соединение если было
	if oldChannel != nil {
		close(oldChannel)
		logger.Info("closed old connection", zap.Int64("chat_id", chatId), zap.Int64("user_id", userId))
	}

	logger.Info("subscriber connected",
		zap.Int64("chat_id", chatId),
		zap.Int64("user_id", userId),
		zap.String("username", username),
		zap.Int("online_count", room.GetOnlineUsersCount()),
	)

	// Отправляем историю сообщений
	go s.sendRecentMessages(ctx, chatId, msgChan)

	// Уведомляем всех об обновлении онлайн пользователей
	go room.BroadcastOnlineUsers()

	return msgChan, nil
}
