package service

import (
	"context"
	"fmt"
	"time"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"github.com/GolZrd/micro-chat/chat-server/internal/repository"
	"go.uber.org/zap"
)

// SendMessage сохраняет сообщение и рассылает подписчикам
func (s *service) SendMessage(ctx context.Context, msg SendMessageDTO) error {
	input := repository.MessageCreateDTO{
		ChatId:       msg.ChatId,
		UserId:       msg.UserId,
		FromUsername: msg.FromUsername,
		Text:         msg.Text,
	}

	logger.Info("sending message", zap.Int64("chat_id", msg.ChatId), zap.String("sent by", msg.FromUsername))

	// Сохраняем сообщение в БД
	err := s.ChatRepository.SendMessage(ctx, input)
	if err != nil {
		logger.Error("failed to save message", zap.Int64("chat_id", msg.ChatId), zap.String("sent by", msg.FromUsername), zap.Error(err))
		return fmt.Errorf("database: failed to save message: %w", err)
	}

	// Создаем DTO для рассылки
	msgDTO := MessageDTO{
		Type:      MessageTypeText,
		From:      msg.FromUsername,
		Text:      msg.Text,
		CreatedAt: time.Now(),
	}

	// Отправляем всем подписчикам сообщение если комната существует
	room := s.getRoom(msg.ChatId)
	if room != nil {
		room.BroadcastMessage(msgDTO)
	}

	return nil
}

// Функция для отправки последних сообщений
func (s *service) sendRecentMessages(ctx context.Context, chatID int64, msgChan chan<- MessageDTO) {
	// Если вдруг канал будет закрыт, то нужно обработать панику
	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic in sendRecentMessages", zap.Int64("chat_id", chatID), zap.Any("panic", r))
		}
	}()

	logger.Debug("loading chat history", zap.Int64("chat_id", chatID))

	// Получаем последние 50 сообщений
	messages, err := s.ChatRepository.RecentMessages(ctx, chatID, 50)
	if err != nil {
		logger.Error("failed to load chat history", zap.Int64("chat_id", chatID), zap.Error(err))

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
			Type:      MessageTypeText,
			From:      msg.From,
			Text:      msg.Text,
			CreatedAt: msg.CreatedAt,
		}:
		case <-ctx.Done():
			logger.Debug("history sending interrupted", zap.Int64("chat_id", chatID))
			return
		}
	}

	logger.Debug("history sent successfully", zap.Int64("chat_id", chatID), zap.Int("message_count", len(messages)))
}
