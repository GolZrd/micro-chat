package model

import "time"

type Chats struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}
