package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	userRepository "github.com/GolZrd/micro-chat/auth/internal/repository/user"
	"go.uber.org/zap"
)

func (s *service) Create(ctx context.Context, info CreateUserDTO) (int64, error) {
	if info.Password != info.PasswordConfirm {
		return 0, errors.New("passwords do not match")
	}

	// Нужно привести email к нижнему регистру
	email := strings.ToLower(info.Email)

	// service DTO → repository DTO
	params := userRepository.CreateUserDTO{
		Username: strings.TrimSpace(info.Username),
		Email:    strings.TrimSpace(email),
		Password: info.Password,
		Role:     info.Role,
	}

	// Мы должны проверить, что пользователя с таким email еще нет
	exists, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		// Проверяем, что это именно "не найден", а не ошибка БД
		if !errors.Is(err, userRepository.ErrUserNotFound) {
			logger.Error("Failed to check if user exists",
				zap.String("email", email),
				zap.Error(err),
			)
			return 0, fmt.Errorf("failed to get user by email: %w", err)
		}
		// Пользователь не найден - продолжаем создание
	} else {
		// Пользователь найден
		logger.Debug("user already exists",
			zap.String("email", email),
		)
		return 0, errors.New("user already exists")
	}

	if exists != nil {
		logger.Debug("attempt to create duplicate user", zap.String("email", email))
		return 0, errors.New("user already exists")
	}

	// Пароль в открытом виде, так как не продакшн, но мы должны его хешировать

	id, err := s.userRepository.Create(ctx, params)
	if err != nil {
		logger.Error("failed to create user",
			zap.String("email", email),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info(
		"user created successfully",
		zap.Int64("user_id", id),
		zap.String("role", info.Role),
	)

	return id, nil
}
