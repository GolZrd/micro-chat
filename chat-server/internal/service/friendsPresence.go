package service

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

func (s *service) FriendsPresence(ctx context.Context, userIds []int64) ([]FriendPresenceDTO, error) {

	presences, err := s.PresenceRepository.GetPresence(ctx, userIds)
	if err != nil {
		logger.Error("Failed to get presence", zap.Error(err))
		return nil, fmt.Errorf("failed to get presence: %w", err)
	}

	result := make([]FriendPresenceDTO, 0, len(presences))
	for _, p := range presences {
		result = append(result, FriendPresenceDTO{
			UserId:     p.UserId,
			IsOnline:   p.IsOnline,
			LastSeenAt: p.LastSeenAt,
		})
	}
	return result, nil
}
