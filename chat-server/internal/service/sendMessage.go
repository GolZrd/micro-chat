package service

import (
	"context"
	"errors"

	"github.com/GolZrd/micro-chat/chat-server/internal/repository"
)

// SendMessage сохраняет сообщение и рассылает подписчикам
func (s *service) SendMessage(ctx context.Context, msg SendMessageDTO) error {
	// Добавим проверку что chat_id существует и отправитель есть в этом чате
	if msg.Chat_id <= 0 {
		return errors.New("chat_id cannot be empty")
	}

	if msg.From_username == "" || msg.Text == "" {
		return errors.New("from and text cannot be empty")
	}

	input := repository.MessageCreateDTO{
		Chat_id:       msg.Chat_id,
		From_username: msg.From_username,
		Text:          msg.Text,
		Created_at:    msg.Created_at,
	}

	// Сохраняем сообщение в БД
	err := s.ChatRepository.SendMessage(ctx, input)
	if err != nil {
		return err
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
				// Канал переполнен, пропускаем
			}

		}
	}
}
