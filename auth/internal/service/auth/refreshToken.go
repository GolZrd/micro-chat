package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/model"
	"go.uber.org/zap"
)

// Исходим из того, что на стороне клиента следят за сроком действия токена, поэтому токен всегда валиден
func (s *service) RefreshToken(ctx context.Context, oldRefreshToken string) (refreshToken string, err error) {
	//Проверяем что токен валиден
	userData, err := s.jwtManager.VerifyToken(oldRefreshToken, []byte(s.RefreshSecretKey))
	if err != nil {
		logger.Warn("Invalid refresh token", zap.String("refresh_token", oldRefreshToken[:8]), zap.Error(err))

		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Если токен нормальный, тогда нам нужно его просрочить и сгенерировать новый токен, после этого записать его в БД
	// Помечаем старый токен как revoked
	err = s.authRepository.RevokeToken(ctx, oldRefreshToken)
	if err != nil {
		// Это не критичный метод, так как либо при первом входе, либо при обновлении токена все токены будут помечены как revoked
		logger.Warn("failed to revoke existing tokens", zap.Int64("user_id", userData.UID), zap.Error(err))
	}

	// Создаем новый токен
	refreshToken, err = s.jwtManager.GenerateToken(model.UserAuthData{Id: userData.UID, Username: userData.Username, Role: userData.Role}, s.RefreshSecretKey, s.refreshTTL)
	if err != nil {
		logger.Error("failed to generate refresh token", zap.Int64("user_id", userData.UID), zap.Error(err))
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Сохраняем новый токен в БД
	err = s.authRepository.CreateRefreshToken(ctx, userData.UID, refreshToken, time.Now().Add(s.refreshTTL))
	// TODO: обработать ошибку
	if err != nil {
		logger.Error("failed to save refresh token",
			zap.Int64("user_id", userData.UID),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to save refresh token: %w", err)
	}

	// Возвращаем новый токен
	return refreshToken, nil
}
