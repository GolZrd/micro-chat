package auth

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/model"
	authRepo "github.com/GolZrd/micro-chat/auth/internal/repository/auth"
	"github.com/GolZrd/micro-chat/auth/internal/utils/jwt"
)

func (s *service) Login(ctx context.Context, email string, password string) (refreshToken string, err error) {
	// Вызываем метод user репозитория для получения данных о пользователе по email
	userData, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		// Ошибка на уровне репозитория
		if errors.Is(err, authRepo.ErrUserNotFound) {
			log.Println("User not found")
			return "", errors.New("invalid credentials")
		}

		log.Println("failed to get user by email:", err)

		return "", err
	}

	// Проверяем пароль в упрощенном варианте
	// TODO: Добавить хеширование
	if userData.Password != password {
		return "", errors.New("invalid credentials")
	}

	log.Printf("Login user with id: %d, email: %s, role: %s", userData.Id, userData.Email, userData.Role)

	// Вызываем репозиторий чтобы сохранить токен в БД, после этого просто возвращаем его пользователю и он у себя подсохраняет
	// У нас будет только 1 устройство, поэтому старые refresh мы будем помечать как revoked

	// Начнем с помечания старых токенов как revoked
	err = s.authRepository.RevokeAllByUserId(ctx, userData.Id)
	// TODO: обработать ошибку
	if err != nil {
		return "", err
	}

	// Генерируем новый токен
	token, err := jwt.GenerateToken(model.UserAuthData{Id: userData.Id, Role: userData.Role}, s.RefreshSecretKey, s.refreshTTL)
	if err != nil {
		log.Println("failed to generate token:", err)
		return "", err
	}

	// Сохраняем токен в БД
	err = s.authRepository.CreateRefreshToken(ctx, userData.Id, token, time.Now().Add(s.refreshTTL))
	if err != nil {
		log.Println("failed to create refresh token:", err)
		return "", err
	}

	// Возвращаем токен
	return token, nil
}
