package service

import (
	"context"
	"sync"

	"github.com/GolZrd/micro-chat/chat-server/internal/repository"
)

type ChatService interface {
	Create(ctx context.Context, usernames []string) (int64, error)
	Delete(ctx context.Context, id int64) error
	SendMessage(ctx context.Context, msg SendMessageDTO) error
	ConnectToChat(ctx context.Context, userId int64, chatID int64) (<-chan MessageDTO, error)
	DisconnectFromChat(chatID int64, subscriberId string)
	MyChats(ctx context.Context, username string) ([]ChatInfoDTO, error)
}

type service struct {
	ChatRepository repository.ChatRepository
	subscribers    map[int64]map[string]chan MessageDTO // chat_id -> username -> channel
	subMutex       sync.RWMutex                         // mutex for subscribers
	subIDCounter   int64                                // counter for subscribers
}

func NewService(chatRepository repository.ChatRepository) ChatService {
	return &service{
		ChatRepository: chatRepository,
		subscribers:    make(map[int64]map[string]chan MessageDTO)}
}
