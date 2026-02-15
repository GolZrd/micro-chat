package model

import (
	"database/sql"
	"time"
)

// User — основная модель для репозитория
type User struct {
	Id        int64        `db:"id"`
	Info      UserInfo     `db:""`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
}

// UserInfo — модель для info
type UserInfo struct {
	Username string `db:"username"`
	Email    string `db:"email"`
	Password string `db:"password"`
	Role     string `db:"role"`
}

// UpdateUserInfo — для частичного обновления
type UpdateUserInfo struct {
	Username *string `db:"username"`
	Email    *string `db:"email"`
}

// Для метода GetByEmail, но для squirrel нужны теги, поэтому эта структура тут дублируется
type UserAuthData struct {
	Id       int64  `db:"id"`
	Username string `db:"username"`
	Email    string `db:"email"`
	Password string `db:"password"`
	Role     string `db:"role"`
}

// Для метода CheckUsersExists
type UserShort struct {
	Id       int64  `db:"id"`
	Username string `db:"username"`
}

// Для метода SearchUser
type UserSearchResult struct {
	Id               int64
	Username         string
	FriendshipStatus string
}
