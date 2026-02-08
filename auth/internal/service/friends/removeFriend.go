package friends

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"go.uber.org/zap"
)

func (s *service) RemoveFriend(ctx context.Context, userId, friendId int64) error {
	err := s.friendsRepository.RemoveFriend(ctx, userId, friendId)
	if err != nil {
		logger.Error("Failed to remove friend", zap.Int64("user_id", userId), zap.Int64("friend_id", friendId), zap.Error(err))
		return fmt.Errorf("failed to remove friend %w", err)
	}

	logger.Info("Remove friend", zap.Int64("user_id", userId), zap.Int64("friend_id", friendId))

	return nil
}
