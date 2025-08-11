package service

import (
	"chat-server/internal/repository"
	"context"
)

type ChatService interface {
	Create(ctx context.Context, usernames []string) (int64, error)
	Delete(ctx context.Context, id int64) error
	SendMessage(ctx context.Context, msg SendMessageDTO) error
}

type service struct {
	ChatRepository repository.ChatRepository
}

func NewService(chatRepository repository.ChatRepository) ChatService {
	return &service{ChatRepository: chatRepository}
}
