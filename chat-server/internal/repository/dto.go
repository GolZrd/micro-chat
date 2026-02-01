package repository

import "time"

type MessageCreateDTO struct {
	ChatId       int64
	FromUsername string
	Text         string
	CreatedAt    time.Time
}

type MessageDTO struct {
	Id        int64
	ChatId    int64
	From      string
	Text      string
	CreatedAt time.Time
}

type ChatInfoDTO struct {
	ID        int64
	Name      string
	Usernames []string
	CreatedAt time.Time
}
