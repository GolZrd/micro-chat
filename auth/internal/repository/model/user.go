package model

import (
	"database/sql"
	"time"
)

// User — основная модель (маппится на proto.User)
type User struct {
	Id        int64        `db:"id"`
	Info      UserInfo     `db:""`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
}

// UserInfo — модель для info (маппится на proto.UserInfo, без password_confirm — оно для валидации)
type UserInfo struct {
	Name     string `db:"name"`
	Email    string `db:"email"`
	Password string `db:"password"`
	Role     string `db:"role"`
}

// UpdateUserInfo — для частичного обновления
type UpdateUserInfo struct {
	Name  *string `db:"name"`
	Email *string `db:"email"`
}
