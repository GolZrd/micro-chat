package model

import (
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	jwt.RegisteredClaims
	UID      int64  `json:"uid"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
