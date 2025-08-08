package model

import "time"

type Message struct {
	Id           int64     `db:"id"`
	ChatId       int64     `db:"chat_id"`
	FromUsername string    `db:"username"`
	Text         string    `db:"text"`
	CreatedAt    time.Time `db:"created_at"`
}
