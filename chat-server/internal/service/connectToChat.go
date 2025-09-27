package service

import (
	"context"
	"fmt"
)

// ConnectToChat подписывает на получение сообщений из чата
func (s *service) ConnectToChat(ctx context.Context, chatID int64) (<-chan MessageDTO, error) {
	// Проверяем что чат существует
	exists, err := s.ChatRepository.ChatExists(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("check chat exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("chat %d not found", chatID)
	}

	// Создаем канал для подписчика
	msgChan := make(chan MessageDTO, 100)

	// Генерируем уникальный ID подписчика
	s.subMutex.Lock()
	s.subIDCounter++
	subscriberId := fmt.Sprintf("sub_%d_%d", chatID, s.subIDCounter) // Можно просто использовать username подписчика

	if s.subscribers[chatID] == nil {
		s.subscribers[chatID] = make(map[string]chan MessageDTO)
	}

	s.subscribers[chatID][subscriberId] = msgChan
	s.subMutex.Unlock()

	// Запускаем горутину для отслеживания отмены контекста
	go func() {
		<-ctx.Done()
		s.DisconnectFromChat(chatID, subscriberId)
	}()

	// Отправляем последние сообщения
	go s.sendRecentMessages(ctx, chatID, msgChan)

	return msgChan, nil
}

// Функция для отправки последних сообщений
func (s *service) sendRecentMessages(ctx context.Context, chatID int64, msgChan chan<- MessageDTO) {
	// Получаем последние 50 сообщений
	messages, err := s.ChatRepository.RecentMessages(ctx, chatID, 50)
	if err != nil {
		return
	}

	for _, msg := range messages {
		select {
		case msgChan <- MessageDTO{
			From:      msg.From,
			Text:      msg.Text,
			CreatedAt: msg.CreatedAt,
		}:
		case <-ctx.Done():
			return
		}
	}
}
