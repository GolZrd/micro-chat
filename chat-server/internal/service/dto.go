package service

import "time"

type MessageType int

const (
	MessageTypeText        MessageType = 1
	MessageTypeOnlineUsers MessageType = 2
)

type SendMessageDTO struct {
	ChatId       int64
	UserId       int64
	FromUsername string
	Text         string
}

type MessageDTO struct {
	Type        MessageType
	From        string
	Text        string
	CreatedAt   time.Time
	OnlineUsers []OnlineUserDTO // Используется только когда у нас тип MessageTypeOnlineUsers
}

type ChatInfoDTO struct {
	ID        int64
	Name      string
	IsDirect  bool
	Usernames []string
	CreatedAt time.Time
}

// OnlineUserDTO информация о подключенном пользователе
type OnlineUserDTO struct {
	UserId   int64  `json:"userId"`
	Username string `json:"username"`
}
