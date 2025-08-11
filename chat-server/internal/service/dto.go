package service

import "time"

type SendMessageDTO struct {
	Chat_id       int64
	From_username string
	Text          string
	Created_at    time.Time
}
