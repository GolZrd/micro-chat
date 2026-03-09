package utils

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// UserClaims — данные из токена
type UserClaims struct {
	UserId   int64
	Username string
}

// ParseTokenClaims — извлекает user_id и username из JWT
// Без проверки подписи — она уже проверена middleware
func ParseTokenClaims(tokenStr string) (*UserClaims, error) {
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	result := &UserClaims{}

	// user_id — пробуем разные ключи
	for _, key := range []string{"uid", "user_id", "sub"} {
		if val, ok := claims[key]; ok {
			switch v := val.(type) {
			case float64:
				result.UserId = int64(v)
			case int64:
				result.UserId = v
			}
			if result.UserId > 0 {
				break
			}
		}
	}

	if result.UserId == 0 {
		return nil, fmt.Errorf("user_id not found in token")
	}

	// username
	for _, key := range []string{"username", "name", "preferred_username"} {
		if val, ok := claims[key].(string); ok && val != "" {
			result.Username = val
			break
		}
	}

	return result, nil
}
