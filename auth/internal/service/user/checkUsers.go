package user

import (
	"context"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"go.uber.org/zap"
)

func (s *service) CheckUsersExists(ctx context.Context, usernames []string) ([]string, error) {
	existingUsers, err := s.userRepository.GetByUsernames(ctx, usernames)
	if err != nil {
		logger.Error("Failed to check users", zap.Strings("usernames", usernames), zap.Error(err))
		return nil, err
	}

	var notFoundUsers []string

	// Используем мапу, для проверки существования пользователя
	existingMap := make(map[string]struct{})
	for _, username := range existingUsers {
		existingMap[username] = struct{}{}
	}

	for _, username := range usernames {
		if _, ok := existingMap[username]; !ok {
			notFoundUsers = append(notFoundUsers, username)
		}
	}

	logger.Info("Checked users exists", zap.Strings("usernames", usernames), zap.Strings("not_found", notFoundUsers))

	return notFoundUsers, nil
}
