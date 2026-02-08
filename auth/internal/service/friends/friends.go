package friends

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/repository/friends/model"
	"go.uber.org/zap"
)

func (s *service) Friends(ctx context.Context, userid int64) ([]model.Friend, error) {
	frineds, err := s.friendsRepository.Friends(ctx, userid)
	if err != nil {
		logger.Error("Failed to get friends", zap.Int64("user_id", userid), zap.Error(err))
		return nil, fmt.Errorf("failed to get friends %w", err)
	}

	logger.Info("Get friends", zap.Int64("user_id", userid), zap.Int("count", len(frineds)))

	return frineds, nil
}
