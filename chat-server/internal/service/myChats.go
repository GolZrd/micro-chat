package service

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

// MyChats возвращает список чатов пользователя
func (s *service) MyChats(ctx context.Context, userId int64) ([]ChatInfoDTO, error) {
	chats, err := s.ChatRepository.UserChats(ctx, userId)
	if err != nil {
		logger.Error("failed to get user chats", zap.Int64("user_id", userId), zap.Error(err))
		return nil, fmt.Errorf("failed to get user chats: %w", err)
	}

	allChats := make([]ChatInfoDTO, 0, len(chats))

	for _, chat := range chats {
		usernames := make([]string, 0, len(chat.Members))
		for _, member := range chat.Members {
			usernames = append(usernames, member.Username)
		}

		allChats = append(allChats, ChatInfoDTO{
			ID:        chat.ID,
			Name:      chat.Name,
			IsDirect:  chat.IsDirect,
			Usernames: usernames,
			CreatedAt: chat.CreatedAt,
		})
	}
	return allChats, nil
}
