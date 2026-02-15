package user

import (
	"context"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/model"
	"go.uber.org/zap"
)

func (s *service) CheckUsersExists(ctx context.Context, usernames []string) (foundUsers []model.UserShort, notFoundUsers []string, err error) {

	foundUsers, err = s.userRepository.GetByUsernames(ctx, usernames)
	if err != nil {
		logger.Error("Failed to check users", zap.Strings("usernames", usernames), zap.Error(err))
		return nil, nil, err
	}

	// Используем мапу, для сборки существующих пользователей
	existingMap := make(map[string]struct{}, len(foundUsers))
	for _, user := range foundUsers {
		existingMap[user.Username] = struct{}{}
	}

	// Собираем несуществующих пользователей
	for _, username := range usernames {
		if _, ok := existingMap[username]; !ok {
			notFoundUsers = append(notFoundUsers, username)
		}
	}

	logger.Info("Checked users exists", zap.Strings("usernames", usernames), zap.Int("found", len(foundUsers)), zap.Strings("not_found", notFoundUsers))

	return foundUsers, notFoundUsers, nil
}
