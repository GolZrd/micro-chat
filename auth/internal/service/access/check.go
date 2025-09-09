package access

import (
	"auth/internal/utils/jwt"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *service) Check(ctx context.Context, accessToken string, endPoint string) error {
	// Верифицируем токен
	claims, err := jwt.VerifyToken(accessToken, []byte(s.AccessSecretKey))
	if err != nil {
		return status.Errorf(codes.Aborted, "invalid access token")
	}

	// Получаем доступные роли для конретного эндпоинта
	allowedRoles, err := s.accessRepository.EndPointRoles(ctx, endPoint)
	if err != nil {
		return status.Error(codes.Internal, "failed to get endpoint roles")
	}

	// Проверяем, что allowedRoles не пустой, если пустой, то в нашем случае все имеют доступ
	if len(allowedRoles) == 0 {
		return nil
	}

	// Проверяем, что у пользователя есть доступ к конкретному эндпоинту
	_, ok := allowedRoles[claims.Role]
	if !ok {
		return status.Errorf(codes.PermissionDenied, "user doesn't have access to endpoint")
	}

	// Если все ок, то возвращаем nil
	return nil
}
