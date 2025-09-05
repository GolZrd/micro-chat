package auth

import (
	"auth/internal/model"
	"auth/internal/utils/jwt"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *service) AccessToken(ctx context.Context, refreshToken string) (accessToken string, err error) {
	// Сначала проверяем валидность токена
	userData, err := jwt.VerifyToken(refreshToken, []byte(s.RefreshSecretKey))
	if err != nil {
		// TODO: Возможно стоит добавить доп. проверки токена на разные ошибки
		return "", status.Errorf(codes.Aborted, "invalid refresh token")
	}
	// Если токен валиден, генерируем новый access токен
	accessToken, err = jwt.GenerateToken(model.UserAuthData{Id: userData.UID, Role: userData.Role}, s.AccessSecretKey, s.accessTTL)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to generate access token")
	}

	// И просто возвращаем access токен
	return accessToken, nil
}
