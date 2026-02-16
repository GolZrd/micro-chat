package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TokenClaims struct {
	jwt.RegisteredClaims
	UID      int64  `json:"uid"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type UserClaims struct {
	UID      int64
	Username string
}

// GetUsernameFromContext извлекает username из JWT токена переданного в контексте
func GetUsernameFromContext(ctx context.Context) (string, error) {
	// Извлекаем токен из метадаты
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization header is not provided")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")

	// Парсим токен без проверки, так как проверка уже прошла в интерцепторе
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, &TokenClaims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := parsedToken.Claims.(*TokenClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	return claims.Username, nil
}

func GetUIDFromContext(ctx context.Context) (int64, error) {
	// Извлекаем токен из метадаты
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return 0, status.Error(codes.Unauthenticated, "authorization header is not provided")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")

	// Парсим токен без проверки, так как проверка уже прошла в интерцепторе
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, &TokenClaims{})
	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := parsedToken.Claims.(*TokenClaims)
	if !ok {
		return 0, fmt.Errorf("invalid token claims")
	}

	return claims.UID, nil
}

func GetUserClaimsFromContext(ctx context.Context) (*UserClaims, error) {
	// Извлекаем токен из метадаты
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization header is not provided")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")

	// Парсим токен без проверки, так как проверка уже прошла в интерцепторе
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, &TokenClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := parsedToken.Claims.(*TokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &UserClaims{UID: claims.UID, Username: claims.Username}, nil
}
