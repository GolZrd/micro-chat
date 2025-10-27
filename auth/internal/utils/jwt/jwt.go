package jwt

import (
	"fmt"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

// Функция для генерации токена. Передаем информацию о пользователе, секретный ключ и время жизни
func GenerateToken(user model.UserAuthData, secretKey string, ttl time.Duration) (string, error) {
	// Добавляем в токен информацию о пользователе
	claims := model.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
		UID:  user.Id,
		Name: user.Name,
		Role: user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

// Функция для проверки токена на валидность
func VerifyToken(tokenStr string, secretKey []byte) (*model.UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &model.UserClaims{}, func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, errors.Errorf("invalid token: %s", err.Error())
	}

	claims, ok := token.Claims.(*model.UserClaims)
	if !ok {
		return nil, errors.Errorf("invalid token claims")
	}

	return claims, nil
}
