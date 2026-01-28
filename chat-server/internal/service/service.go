package service

import (
	"context"
	"sync"

	"github.com/GolZrd/micro-chat/chat-server/internal/client/grpc/auth"
	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"github.com/GolZrd/micro-chat/chat-server/internal/repository"
	"go.uber.org/zap"
)

type ChatService interface {
	// Управление чатами
	Create(ctx context.Context, creatorUsername string, usernames []string) (int64, error)
	Delete(ctx context.Context, id int64) error
	MyChats(ctx context.Context, username string) ([]ChatInfoDTO, error)

	// Сообщения
	SendMessage(ctx context.Context, msg SendMessageDTO) error
	// Подключение к чату
	ConnectToChat(ctx context.Context, userId int64, username string, chatID int64) (<-chan MessageDTO, error)
	DisconnectFromChat(chatId int64, userId int64)

	// Онлайн статусы
	OnlineUsers(chatID int64) []OnlineUserDTO
	OnlineCount(chatID int64) int
	IsUserOnline(chatID int64, userID int64) bool
}

type service struct {
	ChatRepository repository.ChatRepository
	authClient     *auth.Client

	rooms   map[int64]*ChatRoom // chat_id → ChatRoom
	roomsMu sync.RWMutex        // мьютекс для создания и удаления комнат
}

func NewService(chatRepository repository.ChatRepository, authClient *auth.Client) ChatService {
	return &service{
		ChatRepository: chatRepository,
		authClient:     authClient,
		rooms:          make(map[int64]*ChatRoom),
	}
}

// getRoom возвращает комнату по chat_id без создания
func (s *service) getRoom(chatId int64) *ChatRoom {
	s.roomsMu.RLock()
	defer s.roomsMu.RUnlock()
	return s.rooms[chatId]
}

// getOrCreateRoom возвращает комнату по chat_id или создает новую
func (s *service) getOrCreateRoom(chatId int64) *ChatRoom {
	// Проверяем, если комната есть, то возвращаем
	s.roomsMu.RLock()
	room, exists := s.rooms[chatId]
	s.roomsMu.RUnlock()

	if exists {
		return room
	}

	// Если комната не существует, то создаем
	s.roomsMu.Lock()
	defer s.roomsMu.Unlock()

	// Перепроверяем, что комнату еще не создали
	if room, exists := s.rooms[chatId]; exists {
		return room
	}

	room = newChatRoom()
	s.rooms[chatId] = room

	logger.Debug("chat room created", zap.Int64("chat_id", chatId))

	return room
}

func (s *service) deleteRoomIfEmpty(chatId int64) {
	s.roomsMu.Lock()
	defer s.roomsMu.Unlock()

	room, exists := s.rooms[chatId]
	if !exists {
		return
	}

	if room.IsEmpty() {
		delete(s.rooms, chatId)
		logger.Debug("chat room deleted", zap.Int64("chat_id", chatId))
	}
}
