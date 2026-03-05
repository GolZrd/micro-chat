package service

import "time"

const (
	MessageTypeText        = 0
	MessageTypeOnlineUsers = 1
	MessageTypeVoice       = 2
)

type SendMessageDTO struct {
	ChatId        int64
	UserId        int64
	FromUsername  string
	Text          string
	MessageType   int32
	VoiceDuration float32
}

type MessageDTO struct {
	MessageType   int32
	From          string
	Text          string
	CreatedAt     time.Time
	OnlineUsers   []OnlineUserDTO // Используется только когда у нас тип MessageTypeOnlineUsers
	VoiceDuration float32
}

type ChatInfoDTO struct {
	ID        int64
	Name      string
	Usernames []string
	IsDirect  bool
	IsPublic  bool
	CreatorId int64
	CreatedAt time.Time
}

// OnlineUserDTO информация о подключенном пользователе
type OnlineUserDTO struct {
	UserId   int64  `json:"userId"`
	Username string `json:"username"`
}

type FriendPresenceDTO struct {
	UserId     int64
	IsOnline   bool
	LastSeenAt time.Time
}

type PublicChatDTO struct {
	Id          int64
	Name        string
	MemberCount int
	CreatorName string
	CreatedAt   time.Time
}
