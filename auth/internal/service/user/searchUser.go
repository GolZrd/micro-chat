package user

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/model"
	"go.uber.org/zap"
)

func (s *service) SearchUser(ctx context.Context, searchQuery string, currentUserId int64, limit int) ([]model.UserSearchResult, error) {
	if len(searchQuery) < 2 {
		logger.Warn("Query too short", zap.String("query", searchQuery), zap.Int64("current_user_id", currentUserId), zap.Int("limit", limit))
		return nil, fmt.Errorf("query too short")
	}

	users, err := s.userRepository.SearchUser(ctx, searchQuery, currentUserId, limit)
	if err != nil {
		logger.Error("Failed to search users", zap.String("query", searchQuery), zap.Int64("current_user_id", currentUserId), zap.Int("limit", limit), zap.Error(err))
		return nil, fmt.Errorf("failed to search users %w", err)
	}

	return users, nil
}
