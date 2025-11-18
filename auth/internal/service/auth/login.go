package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/model"
	authRepo "github.com/GolZrd/micro-chat/auth/internal/repository/auth"
	"go.uber.org/zap"
)

func (s *service) Login(ctx context.Context, email string, password string) (refreshToken string, userId int64, err error) {
	// Приводим email к нижнему регистру
	lowerEmail := strings.ToLower(email)

	// В проде email в открытом виде передавать не стоит
	logger.Debug("login attempt", zap.String("email", email))

	// Вызываем метод user репозитория для получения данных о пользователе по email
	userData, err := s.userRepository.GetByEmail(ctx, lowerEmail)
	if err != nil {
		// Ошибка на уровне репозитория
		if errors.Is(err, authRepo.ErrUserNotFound) {
			logger.Debug("login failed: user not found", zap.String("email", email))
			return "", 0, errors.New("invalid credentials")
		}

		// Уровень Error в логах, потому что ошибка в БД
		logger.Error("failed to get user by email", zap.Error(err))

		return "", 0, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Проверяем пароль в упрощенном варианте
	// TODO: Добавить хеширование
	if userData.Password != password {
		logger.Debug("login failed: invalid password", zap.String("user_id", userData.Name))
		return "", 0, errors.New("invalid credentials")
	}

	// Логгируем успешный аутентификацию
	logger.Info("user login success", zap.Int64("user_id", userData.Id), zap.String("user_role", userData.Role))

	// Вызываем репозиторий чтобы сохранить токен в БД, после этого просто возвращаем его пользователю и он у себя подсохраняет
	// У нас будет только 1 устройство, поэтому старые refresh мы будем помечать как revoked

	// Начнем с помечания старых токенов как revoked
	err = s.authRepository.RevokeAllByUserId(ctx, userData.Id)
	if err != nil {
		// Это не критичный метод, так как либо при первом входе, либо при обновлении токена все токены будут помечены как revoked
		logger.Warn("failed to revoke existing tokens", zap.Int64("user_id", userData.Id), zap.Error(err))
	}

	// Генерируем новый токен
	token, err := s.jwtManager.GenerateToken(model.UserAuthData{Id: userData.Id, Name: userData.Name, Role: userData.Role}, s.RefreshSecretKey, s.refreshTTL)
	if err != nil {
		logger.Error("failed to generate refresh token", zap.Int64("user_id", userData.Id), zap.Error(err))
		return "", 0, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Сохраняем токен в БД
	err = s.authRepository.CreateRefreshToken(ctx, userData.Id, token, time.Now().Add(s.refreshTTL))
	if err != nil {
		logger.Error("failed to save refresh token",
			zap.Int64("user_id", userData.Id),
			zap.Error(err),
		)
		return "", 0, fmt.Errorf("failed to save refresh token: %w", err)
	}

	// Возвращаем токен и id пользователя
	return token, userData.Id, nil
}
