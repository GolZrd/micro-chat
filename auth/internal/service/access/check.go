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
	allowedRoles, err := s.getEndpointRolesCache(ctx, endPoint)
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

// Дополнительная функция для получения ролей из кэша, если данные не нашлись в кэше, то делаем запрос в бд и возвращаем мапу
func (s *service) getEndpointRolesCache(ctx context.Context, endPoint string) (map[string]struct{}, error) {
	// Пробуем получить данные из кэша
	cacheData, err := s.redisCache.GetEndpointRoles(ctx, endPoint)
	if err != nil {
		// Если не удалось получить данные из кэша, то логируем ошибку, и продолжаем выполнение функции
		log.Printf("failed to get endpoint roles from cache: %v", err)
	} else if cacheData != nil {
		// Если данные нашлись в кэше, то возвращаем их
		return cacheData, nil
	}

	// Если данные не нашлись в кэше, то логируем
	log.Printf("endpoint roles not found in cache")

	// Если данные не нашлись в кэше, то делаем запрос в бд и возвращаем мапу
	roles, err := s.accessRepository.EndPointRoles(ctx, endPoint)
	if err != nil {
		return nil, err
	}

	// Теперь нужно сохранить данные в кэш, можно делать это в отдельной горутине, чтобы не блокироваться
	go func() {
		if err := s.redisCache.SetEndpointRoles(context.Background(), endPoint, roles); err != nil {
			log.Printf("failed to set endpoint roles in cache: %v", err)
		}
	}()

	return roles, nil
}
