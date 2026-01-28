package service

import (
	"sync"
	"time"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

// ChatRoom представляет собой комнату чата с своим мьютексом
// Основная причина использовать chatRoom это отдельный мьютекс на каждую комнату, то есть для каждого чата свой
type ChatRoom struct {
	subscribers map[int64]*Subscriber
	mu          sync.RWMutex
}

// newChatRoom создает новую комнату
func newChatRoom() *ChatRoom {
	return &ChatRoom{subscribers: make(map[int64]*Subscriber)}
}

// AddSubscriber добавляет подписчика в комнату и возвращает канал для закрытия, если до этого было соединение
// Закрывать старый канал нужно для того, чтобы сообщения отправлялись только в один канал.
func (r *ChatRoom) AddSubscriber(sub *Subscriber) chan MessageDTO {
	r.mu.Lock()
	defer r.mu.Unlock()

	var channel chan MessageDTO

	// Проверяем, если подписчик уже есть в комнате, то возвращаем его канал, иначе создаем новый
	if oldSub, exists := r.subscribers[sub.UserId]; exists {
		channel = oldSub.Channel
	}

	r.subscribers[sub.UserId] = sub

	return channel
}

// RemoveSubscriber удаляет подписчика из комнаты
// Возвращает канал для закрытия, если до этого было соединение и флаг пустая ли комната
func (r *ChatRoom) RemoveSubscriber(userId int64) (chan MessageDTO, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sub, exists := r.subscribers[userId]
	if !exists {
		return nil, len(r.subscribers) == 0
	}

	delete(r.subscribers, userId)

	return sub.Channel, len(r.subscribers) == 0
}

// GetOnlineUsers возвращает список онлайн пользователей, их Id и username
func (r *ChatRoom) GetOnlineUsers() []OnlineUserDTO {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]OnlineUserDTO, 0, len(r.subscribers))
	for _, sub := range r.subscribers {
		users = append(users, OnlineUserDTO{
			UserId:   sub.UserId,
			Username: sub.Username,
		})
	}
	return users
}

// GetOnlineUsersCount возвращает количество онлайн пользователей
func (r *ChatRoom) GetOnlineUsersCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.subscribers)
}

// IsUserOnline проверяет по Id, онлайн ли пользователь
func (r *ChatRoom) IsUserOnline(userId int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.subscribers[userId]
	return exists
}

// IsEmpty проверяет, пустая ли комната
func (r *ChatRoom) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.subscribers) == 0
}

// BroadcastMessage рассылает сообщение всем подписчикам в комнате
func (r *ChatRoom) BroadcastMessage(msg MessageDTO) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, sub := range r.subscribers {
		select {
		case sub.Channel <- msg:
		default:
			// Канал переполнен, пропускаем
			logger.Warn("failed to deliver message - channel full, skip message", zap.Int64("user_id", sub.UserId))
		}
	}
}

func (r *ChatRoom) BroadcastOnlineUsers() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.subscribers) == 0 {
		return
	}

	// Собираем список онлайн пользователей
	onlineUsers := make([]OnlineUserDTO, 0, len(r.subscribers))
	for _, sub := range r.subscribers {
		onlineUsers = append(onlineUsers, OnlineUserDTO{
			UserId:   sub.UserId,
			Username: sub.Username,
		})
	}

	// Упаковываем в MessageDTO с специальным типом
	msg := MessageDTO{
		Type:        MessageTypeOnlineUsers,
		OnlineUsers: onlineUsers,
		CreatedAt:   time.Now(),
	}

	for _, sub := range r.subscribers {
		select {
		case sub.Channel <- msg:
		default:
			logger.Warn("failed to deliver message - channel full, skip online update", zap.Int64("user_id", sub.UserId))
		}
	}

}
