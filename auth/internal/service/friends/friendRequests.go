package friends

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/repository/friends/model"
	"go.uber.org/zap"
)

func (s *service) FriendRequests(ctx context.Context, userId int64) ([]model.FriendRequest, error) {
	requests, err := s.friendsRepository.FriendRequests(ctx, userId)
	if err != nil {
		logger.Error("Failed to get friend requests", zap.Int64("user_id", userId), zap.Error(err))
		return nil, fmt.Errorf("failed to get friend requests %w", err)
	}

	logger.Info("Get friend requests", zap.Int64("user_id", userId), zap.Int("count", len(requests)))

	return requests, nil
}
