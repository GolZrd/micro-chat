package service

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"github.com/GolZrd/micro-chat/chat-server/internal/repository"
	"go.uber.org/zap"
)

// SendMessage сохраняет сообщение и рассылает подписчикам
func (s *service) SendMessage(ctx context.Context, msg SendMessageDTO) error {
	input := repository.MessageCreateDTO{
		Chat_id:       msg.Chat_id,
		From_username: msg.From_username,
		Text:          msg.Text,
		Created_at:    msg.Created_at,
	}

	logger.Info("sending message", zap.Int64("chat_id", msg.Chat_id), zap.String("sent by", msg.From_username))

	// Сохраняем сообщение в БД
	err := s.ChatRepository.SendMessage(ctx, input)
	if err != nil {
		logger.Error("failed to save message", zap.Int64("chat_id", msg.Chat_id), zap.String("sent by", msg.From_username), zap.Error(err))
		return fmt.Errorf("database: failed to save message: %w", err)
	}

	// Создаем DTO для рассылки
	msgDTO := MessageDTO{
		From:      msg.From_username,
		Text:      msg.Text,
		CreatedAt: msg.Created_at,
	}

	// Рассылаем сообщение подписчикам
	s.broadcastMessage(msg.Chat_id, msgDTO)

	return nil
}

// broadcastMessage рассылает сообщение всем подписчикам
func (s *service) broadcastMessage(chatID int64, msg MessageDTO) {
	s.subMutex.RLock()
	defer s.subMutex.RUnlock()

	if chatSubs, exists := s.subscribers[chatID]; exists {
		for _, ch := range chatSubs {
			select {
			case ch <- msg:
				// Отправлено успешно
			default:
				logger.Debug("failed to deliver message - channel full", zap.String("sent by", msg.From), zap.Int64("chat_id", chatID))
				// Канал переполнен, пропускаем
			}

		}
	}
}
