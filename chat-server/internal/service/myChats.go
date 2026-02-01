package service

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

// MyChats возвращает список чатов пользователя
func (s *service) MyChats(ctx context.Context, username string) ([]ChatInfoDTO, error) {
	chats, err := s.ChatRepository.UserChats(ctx, username)
	if err != nil {
		logger.Error("failed to get user chats", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("failed to get user chats: %w", err)
	}

	allChats := make([]ChatInfoDTO, 0, len(chats))

	for _, chat := range chats {
		allChats = append(allChats, ChatInfoDTO{
			ID:        chat.ID,
			Name:      chat.Name,
			Usernames: chat.Usernames,
			CreatedAt: chat.CreatedAt,
		})
	}
	return allChats, nil
}
