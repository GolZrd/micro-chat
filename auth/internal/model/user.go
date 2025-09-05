package model

import (
	"database/sql"
	"time"
)

// Общая доменная модель для всех слоев, это то, что будет возвращаться от других слоев
type User struct {
	Id        int64
	Info      UserInfo
	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

// поле PasswordConfirm здесь не нужно
type UserInfo struct {
	Name     string
	Email    string
	Password string
	Role     string
}

type UpdateUserInfo struct {
	Name  *string `db:"name"`
	Email *string `db:"email"`
}

// Модель для получения пользователя по email
type UserAuthData struct {
	Id       int64
	Email    string
	Password string
	Role     string
}
