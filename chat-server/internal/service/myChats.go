package service

import (
	"context"
	"fmt"
)

// MyChats возвращает список чатов пользователя
func (s *service) MyChats(ctx context.Context, username string) ([]ChatInfoDTO, error) {
	chats, err := s.ChatRepository.UserChats(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user chats: %w", err)
	}

	allChats := make([]ChatInfoDTO, 0, len(chats))

	for _, chat := range chats {
		allChats = append(allChats, ChatInfoDTO{
			ID:        chat.ID,
			Usernames: chat.Usernames,
			CreatedAt: chat.CreatedAt,
		})
	}
	return allChats, nil
}
