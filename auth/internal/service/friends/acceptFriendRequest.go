package friends

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"go.uber.org/zap"
)

func (s *service) AcceptFriendRequest(ctx context.Context, requestId int64, userId int64) error {
	err := s.friendsRepository.AcceptFriendRequest(ctx, requestId, userId)
	if err != nil {
		logger.Error("Failed to accept friend request", zap.Int64("request_id", requestId), zap.Int64("user_id", userId), zap.Error(err))
		return fmt.Errorf("failed to accept friend request %w", err)
	}

	logger.Info("Accept friend request", zap.Int64("request_id", requestId), zap.Int64("user_id", userId))

	return nil
}
