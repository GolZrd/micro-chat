package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

// ConnectToChat подписывает на получение сообщений из чата
// Может следует все таки добавить username подписчика, чтобы было нагляднее
func (s *service) ConnectToChat(ctx context.Context, userId int64, chatID int64) (<-chan MessageDTO, error) {

	logger.Info("Connecting to chat", zap.Int64("user_id", userId), zap.Int64("chat_id", chatID))

	// Проверяем что чат существует
	exists, err := s.ChatRepository.ChatExists(ctx, chatID)
	if err != nil {
		logger.Error("Failed to check chat exists", zap.Int64("chat_id", chatID), zap.Error(err))
		return nil, fmt.Errorf("check chat exists: %w", err)
	}
	if !exists {
		logger.Warn("chat not found", zap.Int64("chat_id", chatID))
		return nil, fmt.Errorf("chat %d not found", chatID)
	}

	// Создаем канал для подписчика
	msgChan := make(chan MessageDTO, 100)

	// Генерируем уникальный ID подписчика
	s.subMutex.Lock()
	s.subIDCounter++
	subscriberId := fmt.Sprintf("sub_%d_%d", chatID, userId) // используем просто ID чата и ID пользователя

	if s.subscribers[chatID] == nil {
		s.subscribers[chatID] = make(map[string]chan MessageDTO)
	}

	s.subscribers[chatID][subscriberId] = msgChan
	s.subMutex.Unlock()

	logger.Info("subscriber connected",
		zap.Int64("chat_id", chatID),
		zap.String("subscriber_id", subscriberId),
		zap.Int64("total_subscribers", s.subIDCounter),
	)

	// Запускаем горутину для отслеживания отмены контекста
	go func() {
		<-ctx.Done()

		// Логгируем причину отключения
		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				logger.Debug("subscriber disconected", zap.String("subsriber_id", subscriberId), zap.String("reason", "canceld"))
			} else if errors.Is(err, context.DeadlineExceeded) {
				logger.Warn("subscriber timeout", zap.String("subsriber_id", subscriberId))
			}
		}

		s.DisconnectFromChat(chatID, subscriberId)
	}()

	// Отправляем последние сообщения
	go s.sendRecentMessages(ctx, chatID, subscriberId, msgChan)

	return msgChan, nil
}

// Функция для отправки последних сообщений
func (s *service) sendRecentMessages(ctx context.Context, chatID int64, subscriberId string, msgChan chan<- MessageDTO) {
	// Если вдруг канал будет закрыт, то нужно обработать панику
	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic in sendRecentMessages", zap.Int64("chat_id", chatID), zap.String("subscriber_id", subscriberId), zap.Any("panic", r))
		}
	}()

	logger.Debug("loading chat history", zap.Int64("chat_id", chatID), zap.String("subscriber_id", subscriberId))

	// Получаем последние 50 сообщений
	messages, err := s.ChatRepository.RecentMessages(ctx, chatID, 50)
	if err != nil {
		logger.Error("failed to load chat history", zap.Int64("chat_id", chatID), zap.String("subscriber_id", subscriberId), zap.Error(err))

		// Отправляем сообщение об ошибке
		select {
		case msgChan <- MessageDTO{
			From:      "system",
			Text:      "Failed to load chat history",
			CreatedAt: time.Now(),
		}:
		case <-ctx.Done():
			return
		}
		return
	}

	// Если сообщений нет, то просто логируем
	if len(messages) == 0 {
		logger.Debug("no messages in chat history", zap.Int64("chat_id", chatID))
		return
	}

	// Отправляем сообщения
	for _, msg := range messages {
		select {
		case msgChan <- MessageDTO{
			From:      msg.From,
			Text:      msg.Text,
			CreatedAt: msg.CreatedAt,
		}:
		case <-ctx.Done():
			logger.Debug("history sending interrupted", zap.Int64("chat_id", chatID), zap.String("subscriber_id", subscriberId))
			return
		}
	}

	logger.Debug("history sent successfully", zap.Int64("chat_id", chatID), zap.String("subscriber_id", subscriberId))
}
