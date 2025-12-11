package service

import (
	"context"
	"sync"

	"github.com/GolZrd/micro-chat/chat-server/internal/client/grpc/auth"
	"github.com/GolZrd/micro-chat/chat-server/internal/repository"
)

type ChatService interface {
	Create(ctx context.Context, creatorUsername string, usernames []string) (int64, error)
	Delete(ctx context.Context, id int64) error
	SendMessage(ctx context.Context, msg SendMessageDTO) error
	ConnectToChat(ctx context.Context, userId int64, chatID int64) (<-chan MessageDTO, error)
	DisconnectFromChat(chatID int64, subscriberId string)
	MyChats(ctx context.Context, username string) ([]ChatInfoDTO, error)
}

type service struct {
	ChatRepository repository.ChatRepository
	authClient     *auth.Client
	subscribers    map[int64]map[string]chan MessageDTO // chat_id -> username -> channel
	subMutex       sync.RWMutex                         // mutex for subscribers
	subIDCounter   int64                                // counter for subscribers
}

func NewService(chatRepository repository.ChatRepository, authClient *auth.Client) ChatService {
	return &service{
		ChatRepository: chatRepository,
		authClient:     authClient,
		subscribers:    make(map[int64]map[string]chan MessageDTO)}
}
