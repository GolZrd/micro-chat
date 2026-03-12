package hub

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/GolZrd/micro-chat/web-gateway/internal/logger"
	"go.uber.org/zap"
)

// Notification - уведомление для клиента
type Notification struct {
	Type          string  `json:"type"`
	ChatId        int64   `json:"chat_id"`
	ChatName      string  `json:"chat_name"`
	SenderName    string  `json:"sender_name"`
	SenderId      int64   `json:"sender_id"`
	Text          string  `json:"text"`
	MessageType   string  `json:"message_type"`
	VoiceDuration float32 `json:"voice_duration"`
	Timestamp     int64   `json:"timestamp"`
	FileUrl       string  `json:"file_url"`
	FileName      string  `json:"file_name"`
	FileSize      int64   `json:"file_size"`
}

// connection - одно WS соединение (вкладка)
type connection struct {
	ch chan []byte
}

// Hub - управляет подключениями пользователей
type Hub struct {
	// userID → список соединений (несколько вкладок)
	connections map[int64][]*connection
	mu          sync.RWMutex
}

// NewHub создаёт Hub
func NewHub() *Hub {
	return &Hub{
		connections: make(map[int64][]*connection),
	}
}

// Subscribe - подключает пользователя, возвращает канал
func (h *Hub) Subscribe(userId int64) chan []byte {
	h.mu.Lock()
	defer h.mu.Unlock()

	conn := &connection{
		ch: make(chan []byte, 128),
	}
	h.connections[userId] = append(h.connections[userId], conn)

	logger.Info("user subscribed to notifications",
		zap.Int64("user_id", userId),
		zap.Int("active_connections", len(h.connections[userId])),
	)

	return conn.ch
}

// Unsubscribe - отключает конкретное соединение
func (h *Hub) Unsubscribe(userId int64, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.connections[userId]

	for i, conn := range conns {
		if conn.ch == ch {
			close(conn.ch)
			h.connections[userId] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	if len(h.connections[userId]) == 0 {
		delete(h.connections, userId)
	}

	logger.Info("user unsubscribed from notifications",
		zap.Int64("user_id", userId),
	)
}

// NotifyUser - отправить уведомление одному пользователю
func (h *Hub) NotifyUser(userId int64, notification Notification) {
	notification.Timestamp = time.Now().UnixMilli()

	data, err := json.Marshal(notification)
	if err != nil {
		logger.Error("failed to marshal notification", zap.Error(err))
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.connections[userId]
	if !ok {
		return
	}

	for _, conn := range conns {
		select {
		case conn.ch <- data:
		default:
			logger.Warn("failed to deliver message - channel full, skip notification", zap.Int64("user_id", userId))
		}
	}
}

// NotifyUsers - отправить уведомление списку пользователей, исключая excludeID
func (h *Hub) NotifyUsers(userIds []int64, excludeid int64, notification Notification) {
	notification.Timestamp = time.Now().UnixMilli()

	data, err := json.Marshal(notification)
	if err != nil {
		logger.Error("failed to marshal notification", zap.Error(err))
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, uid := range userIds {
		if uid == excludeid {
			continue
		}

		conns, ok := h.connections[uid]
		if !ok {
			continue
		}

		for _, conn := range conns {
			select {
			case conn.ch <- data:
			default:
				logger.Warn("failed to deliver message - channel full, skip notification", zap.Int64("user_id", uid))
			}
		}
	}
}

// IsOnline - подключён ли пользователь
func (h *Hub) IsOnline(userId int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns, ok := h.connections[userId]
	return ok && len(conns) > 0
}

// OnlineCount - количество подключённых пользователей
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}
