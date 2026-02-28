package repository

import "time"

// CreateChatDTO - DTO для создания чата
type CreateChatDTO struct {
	Name      string
	IsGroup   bool // true = групповой, false = личный
	IsPublic  bool
	CreatorId int64
	Members   []MemberDTO
}

// MessageCreateDTO - DTO для сохранения сообщения
type MessageCreateDTO struct {
	ChatId       int64
	UserId       int64
	FromUsername string
	Text         string
}

// MessageDTO - DTO для получения сообщения
type MessageDTO struct {
	Id        int64
	ChatId    int64
	UserId    int64
	From      string
	Text      string
	CreatedAt time.Time
}

type MemberDTO struct {
	UserId   int64
	Username string
}

// ChatInfoDTO - DTO для получения информации о чате
type ChatInfoDTO struct {
	Id        int64
	Name      string
	IsDirect  bool
	IsPublic  bool
	CreatorId int64
	Members   []MemberDTO
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PublicChatDTO struct {
	Id          int64
	Name        string
	MemberCount int
	CreatorName string
	CreatedAt   time.Time
}
