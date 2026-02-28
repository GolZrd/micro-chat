package service

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

func (s *service) PublicChats(ctx context.Context, search string) ([]PublicChatDTO, error) {
	chats, err := s.ChatRepository.PublicChats(ctx, search)
	if err != nil {
		logger.Error("failed to get public chats", zap.String("search", search), zap.Error(err))
		return nil, fmt.Errorf("get public chats: %w", err)
	}

	res := make([]PublicChatDTO, 0, len(chats))
	for _, chat := range chats {
		res = append(res, PublicChatDTO{
			Id:          chat.Id,
			Name:        chat.Name,
			MemberCount: chat.MemberCount,
			CreatorName: chat.CreatorName,
			CreatedAt:   chat.CreatedAt,
		})
	}

	return res, nil
}
