package jwt

import (
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/model"
)

// JWTManager - интерфейс для работы с JWT
type JWTManager interface {
	GenerateToken(user model.UserAuthData, secretKey string, ttl time.Duration) (string, error)
	VerifyToken(tokenStr string, secretKey []byte) (*model.UserClaims, error)
}

// Структура для работы с токенами
type Manager struct{}

func NewManager() JWTManager {
	return &Manager{}
}
