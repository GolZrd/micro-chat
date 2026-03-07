package service

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

func (s *service) UnreadCounts(ctx context.Context, userId int64) (map[int64]int32, error) {
	res, err := s.UnreadRepository.AllUnreadCounts(ctx, userId)
	if err != nil {
		logger.Error("failed to get unread counts", zap.Int64("user_id", userId), zap.Error(err))
		return nil, fmt.Errorf("get unread counts: %w", err)
	}
	return res, nil
}
