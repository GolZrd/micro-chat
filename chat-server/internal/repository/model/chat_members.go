package model

import "time"

type ChatMembers struct {
	ChatId   int64     `db:"chat_id"`
	Username string    `db:"username"`
	JoinedAt time.Time `db:"joined_at"`
}
