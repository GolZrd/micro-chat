package auth

import (
	"context"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/model"
	"github.com/GolZrd/micro-chat/auth/internal/utils/jwt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Исходим из того, что на стороне клиента следят за сроком действия токена, поэтому токен всегда валиден
func (s *service) RefreshToken(ctx context.Context, oldRefreshToken string) (refreshToken string, err error) {
	//Проверяем что токен валиден
	userData, err := jwt.VerifyToken(oldRefreshToken, []byte(s.RefreshSecretKey))
	if err != nil {
		return "", status.Errorf(codes.Aborted, "invalid refresh token")
	}

	// Если токен нормальный, тогда нам нужно его просрочить и сгенерировать новый токен, после этого записать его в БД
	// Помечаем старый токен как revoked
	err = s.authRepository.RevokeToken(ctx, oldRefreshToken)
	if err != nil {
		// TODO: обработать ошибку
		return "", err
	}

	// Создаем новый токен
	refreshToken, err = jwt.GenerateToken(model.UserAuthData{Id: userData.UID, Name: userData.Name, Role: userData.Role}, s.RefreshSecretKey, s.refreshTTL)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to generate refresh token")
	}

	// Сохраняем новый токен в БД
	err = s.authRepository.CreateRefreshToken(ctx, userData.UID, refreshToken, time.Now().Add(s.refreshTTL))
	// TODO: обработать ошибку
	if err != nil {
		return "", err
	}

	// Возвращаем новый токен
	return refreshToken, nil
}
