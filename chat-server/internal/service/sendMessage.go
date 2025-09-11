package service

import (
	"context"
	"errors"

	"github.com/GolZrd/micro-chat/chat-server/internal/repository"
)

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

	err := s.ChatRepository.SendMessage(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
