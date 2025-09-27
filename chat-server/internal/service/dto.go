package service

import "time"

type SendMessageDTO struct {
	Chat_id       int64
	From_username string
	Text          string
	Created_at    time.Time
}

type MessageDTO struct {
	From      string
	Text      string
	CreatedAt time.Time
}

type ChatInfoDTO struct {
	ID        int64
	Usernames []string
	CreatedAt time.Time
}
