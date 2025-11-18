package auth

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/model"
	"go.uber.org/zap"
)

func (s *service) AccessToken(ctx context.Context, refreshToken string) (accessToken string, err error) {
	// Сначала проверяем валидность токена
	userData, err := s.jwtManager.VerifyToken(refreshToken, []byte(s.RefreshSecretKey))
	if err != nil {
		logger.Warn("Invalid refresh token", zap.String("refresh_token", refreshToken[:8]), zap.Error(err))

		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Используем уровень дебаг, для отладки
	logger.Debug("Token verified",
		zap.String("user", userData.Name),
		zap.String("role", userData.Role),
		zap.String("refresh_token", refreshToken[:8]),
	)

	// Если токен валиден, генерируем новый access токен
	accessToken, err = s.jwtManager.GenerateToken(model.UserAuthData{Id: userData.UID, Name: userData.Name, Role: userData.Role}, s.AccessSecretKey, s.accessTTL)
	if err != nil {
		// Уровень Error в логах
		logger.Error("Failed to generate access token", zap.Error(err))
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// И просто возвращаем access токен
	return accessToken, nil
}
