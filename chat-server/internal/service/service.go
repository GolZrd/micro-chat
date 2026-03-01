package service

import (
	"context"
	"sync"

	"github.com/GolZrd/micro-chat/chat-server/internal/client/grpc/auth"
	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"github.com/GolZrd/micro-chat/chat-server/internal/repository"
	"github.com/GolZrd/micro-chat/chat-server/internal/repository/presence"
	"go.uber.org/zap"
)

type ChatService interface {
	// Управление чатами
	Create(ctx context.Context, name string, isPublic bool, creatorId int64, creatorUsername string, usernames []string) (int64, error)
	Delete(ctx context.Context, id int64) error
	MyChats(ctx context.Context, userId int64) ([]ChatInfoDTO, error)
	GetOrCreateDirectChat(ctx context.Context, currentUserId int64, currentUsername string, targetUserId int64, targetUsername string) (int64, bool, error)
	AddMember(ctx context.Context, chatId int64, userId int64, username string) error
	RemoveMember(ctx context.Context, chatId int64, userId int64, targetUserId int64) error
	JoinChat(ctx context.Context, chatId int64, userId int64, username string) error
	PublicChats(ctx context.Context, search string) ([]PublicChatDTO, error)

	// Сообщения
	SendMessage(ctx context.Context, msg SendMessageDTO) error
	// Подключение к чату
	ConnectToChat(ctx context.Context, userId int64, username string, chatID int64) (<-chan MessageDTO, error)
	DisconnectFromChat(chatId int64, userId int64)

	// Онлайн статусы в чате
	OnlineUsers(chatID int64) []OnlineUserDTO
	OnlineCount(chatID int64) int
	IsUserOnline(chatID int64, userID int64) bool

	// Присутствие
	Heartbeat(ctx context.Context, userId int64) error
	FriendsPresence(ctx context.Context, userIds []int64) ([]FriendPresenceDTO, error)
}

type service struct {
	ChatRepository     repository.ChatRepository
	PresenceRepository presence.RedisRepository
	authClient         *auth.Client

	rooms   map[int64]*ChatRoom // chat_id → ChatRoom
	roomsMu sync.RWMutex        // мьютекс для создания и удаления комнат
}

func NewService(chatRepository repository.ChatRepository, presenceRepo presence.RedisRepository, authClient *auth.Client) ChatService {
	return &service{
		ChatRepository:     chatRepository,
		PresenceRepository: presenceRepo,
		authClient:         authClient,
		rooms:              make(map[int64]*ChatRoom),
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
