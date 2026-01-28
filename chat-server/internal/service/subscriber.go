package service

import "time"

// Subscriber представляет собой подключенного пользователя к чату
type Subscriber struct {
	Channel  chan MessageDTO
	UserId   int64
	Username string
	JoinedAt time.Time
}
