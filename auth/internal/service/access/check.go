package access

import (
	"context"
	"log"

	"github.com/GolZrd/micro-chat/auth/internal/utils/jwt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *service) Check(ctx context.Context, accessToken string, endPoint string) error {
	// Верифицируем токен
	claims, err := jwt.VerifyToken(accessToken, []byte(s.AccessSecretKey))
	if err != nil {
		log.Printf("invalid token: %v", err)
		return status.Errorf(codes.Unauthenticated, "invalid access token")
	}

	// Получаем доступные роли для конретного эндпоинта
	allowedRoles, err := s.accessRepository.EndPointRoles(ctx, endPoint)
	if err != nil {
		log.Printf("failed to get endpoint roles: %v", err)
		return status.Error(codes.Internal, "failed to get endpoint roles")
	}

	log.Printf("user role: %s", claims.Role)

	// Проверяем, что allowedRoles не пустой, если пустой, то в нашем случае все имеют доступ
	if len(allowedRoles) == 0 {
		log.Println("Public endpoint - access granted")
		return nil
	}

	// Проверяем, что у пользователя есть доступ к конкретному эндпоинту
	_, ok := allowedRoles[claims.Role]
	if !ok {
		log.Printf("Access denied for user")
		return status.Errorf(codes.PermissionDenied, "user doesn't have access to endpoint")
	}

	// Если все ок, то возвращаем nil
	return nil
}
