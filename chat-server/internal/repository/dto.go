package repository

import "time"

type MessageCreateDTO struct {
	Chat_id       int64
	From_username string
	Text          string
	Created_at    time.Time
}
