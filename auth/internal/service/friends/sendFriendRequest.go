package friends

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"go.uber.org/zap"
)

func (s *service) SendFriendRequest(ctx context.Context, userId int64, targetUsername string, targetUserId int64) error {
	// Если передан username, то по нему находим id пользователя
	if targetUserId == 0 && targetUsername != "" {
		user, err := s.userRepository.GetByUsername(ctx, targetUsername)
		if err != nil {
			logger.Error("Failed to get user by username", zap.String("username", targetUsername), zap.Error(err))
			return fmt.Errorf("failed to get user by username %w", err)
		}
		targetUserId = user.Id

	}

	if targetUserId == 0 {
		return fmt.Errorf("user not specified")
	}

	if userId == targetUserId {
		return fmt.Errorf("cannot add yourself as friend")
	}

	err := s.friendsRepository.SendFriendRequest(ctx, userId, targetUserId)
	if err != nil {
		logger.Error("Failed to send friend request", zap.Int64("user_id", userId), zap.Int64("target_user_id", targetUserId), zap.Error(err))
		return fmt.Errorf("failed to send friend request %w", err)
	}

	logger.Info("Send friend request", zap.Int64("user_id", userId), zap.Int64("target_user_id", targetUserId))

	return nil
}
