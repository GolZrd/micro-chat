package model

import "time"

type Friend struct {
	Id        int64
	UserId    int64
	Username  string
	AvatarUrl string
}

type FriendRequest struct {
	Id            int64
	FromUserId    int64
	FromUsername  string
	CreatedAt     time.Time
	FromAvatarUrl string
}
