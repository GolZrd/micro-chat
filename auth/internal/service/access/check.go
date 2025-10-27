package access

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/utils/jwt"
	"go.uber.org/zap"
)

func (s *service) Check(ctx context.Context, accessToken string, endPoint string) error {
	// Верифицируем токен
	claims, err := jwt.VerifyToken(accessToken, []byte(s.AccessSecretKey))
	if err != nil {
		logger.Warn("Invalid access token", zap.String("enpoint", endPoint), zap.Error(err))

		return fmt.Errorf("invalid access token: %w", err)
	}

	// Используем уровень дебаг, для отладки
	logger.Debug("Token verified",
		zap.String("user", claims.Name),
		zap.String("role", claims.Role),
		zap.String("endpoint", endPoint),
	)

	// Получаем доступные роли для конретного эндпоинта
	allowedRoles, err := s.getEndpointRolesCache(ctx, endPoint)
	if err != nil {
		// Уровень Error в логах
		logger.Error("Failed to get endpoint roles", zap.String("endpoint", endPoint), zap.Error(err))
		return err
	}

	// Проверяем, что allowedRoles не пустой, если пустой, то в нашем случае все имеют доступ
	if len(allowedRoles) == 0 {
		logger.Debug("Public endpoint detected", zap.String("endpoint", endPoint))
		return nil
	}

	// Проверяем, что у пользователя есть доступ к конкретному эндпоинту
	_, ok := allowedRoles[claims.Role]
	if !ok {
		// Не критическая ошибка, поэтому уровень Warn
		logger.Warn("Permission denied", zap.String("user", claims.Name), zap.String("user_role", claims.Role), zap.Strings("allowedRoles", getRolesList(allowedRoles)), zap.String("endpoint", endPoint))

		return fmt.Errorf("permission_denied: user %s with role %s has no access to %s",
			claims.Name, claims.Role, endPoint)
	}
	// логируем на DEBUG (API слой уже залогировал на INFO)
	logger.Debug("Access granted by service",
		zap.String("user", claims.Name),
		zap.String("role", claims.Role),
		zap.String("endpoint", endPoint),
	)

	// Если все ок, то возвращаем nil
	return nil
}

// Дополнительная функция для получения ролей из кэша, если данные не нашлись в кэше, то делаем запрос в бд и возвращаем мапу
func (s *service) getEndpointRolesCache(ctx context.Context, endPoint string) (map[string]struct{}, error) {
	// Пробуем получить данные из кэша
	cacheData, err := s.redisCache.GetEndpointRoles(ctx, endPoint)
	if err != nil {
		// Если не удалось получить данные из кэша, то логируем ошибку, и продолжаем выполнение функции
		logger.Warn("Failed to get endpoint roles from cache", zap.String("endpoint", endPoint), zap.Error(err))
	} else if cacheData != nil {
		// Debug - для отладки, можно увидеть что данные из кэша
		logger.Debug("Cache hit for endpoint",
			zap.String("endpoint", endPoint),
			zap.Int("roles_count", len(cacheData)),
		)
		return cacheData, nil
	}

	// Логируем cache miss
	logger.Debug("Cache miss, fetching from database", zap.String("endpoint", endPoint))

	// Если данные не нашлись в кэше, то делаем запрос в бд и возвращаем мапу
	roles, err := s.accessRepository.EndPointRoles(ctx, endPoint)
	if err != nil {
		// Если ошибка в БД, то лоигруем Error
		logger.Error("Failed to get endpoint roles from database", zap.String("endpoint", endPoint), zap.Error(err))
		return nil, fmt.Errorf("database_error: %w", err)
	}

	// Теперь нужно сохранить данные в кэш, можно делать это в отдельной горутине, чтобы не блокироваться
	go func() {
		if err := s.redisCache.SetEndpointRoles(context.Background(), endPoint, roles); err != nil {
			// Если не удалось сохранить данные в кэш, то логируем ошибку Warn
			logger.Warn("failed to set endpoint roles in cache", zap.String("endpoint", endPoint), zap.Error(err))
		} else {
			// else только чтобы залогировать успешную запись
			logger.Debug("Set endpoint roles in cache", zap.String("endpoint", endPoint), zap.Int("roles_count", len(roles)))
		}
	}()

	return roles, nil
}

// Helper функция для преобразования map в slice (для логирования)
func getRolesList(roles map[string]struct{}) []string {
	result := make([]string, 0, len(roles))
	for role := range roles {
		result = append(result, role)
	}
	return result
}
