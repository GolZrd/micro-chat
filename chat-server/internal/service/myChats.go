package service

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"github.com/GolZrd/micro-chat/chat-server/internal/repository"
	"go.uber.org/zap"
)

// MyChats возвращает список чатов пользователя
func (s *service) MyChats(ctx context.Context, userId int64) ([]ChatInfoDTO, error) {
	chats, err := s.ChatRepository.UserChats(ctx, userId)
	if err != nil {
		logger.Error("failed to get user chats", zap.Int64("user_id", userId), zap.Error(err))
		return nil, fmt.Errorf("failed to get user chats: %w", err)
	}

	// Получаем непрочитанные
	unreadCounts, err := s.UnreadRepository.AllUnreadCounts(ctx, userId)
	if err != nil {
		logger.Error("failed to get unread counts", zap.Int64("user_id", userId), zap.Error(err))

		// Если не удалось получить непрочитанные, создаем пустую мапу
		unreadCounts = make(map[int64]int32)
	}

	// Собираем id чатов
	chatIds := make([]int64, 0, len(chats))
	for _, chat := range chats {
		chatIds = append(chatIds, chat.Id)
	}

	// Получаем последние сообщения
	lastMessages, err := s.ChatRepository.LastMessages(ctx, chatIds)
	if err != nil {
		logger.Error("failed to get last messages", zap.Int64s("chat_ids", chatIds), zap.Error(err))
		lastMessages = make(map[int64]repository.LastMessageDTO)
	}

	allChats := make([]ChatInfoDTO, 0, len(chats))

	for _, chat := range chats {
		usernames := make([]string, 0, len(chat.Members))
		memberIds := make([]int64, 0, len(chat.Members))
		for _, member := range chat.Members {
			usernames = append(usernames, member.Username)
			memberIds = append(memberIds, member.UserId)
		}

		dto := ChatInfoDTO{
			ID:          chat.Id,
			Name:        chat.Name,
			Usernames:   usernames,
			MemberIds:   memberIds,
			IsDirect:    chat.IsDirect,
			IsPublic:    chat.IsPublic,
			CreatorId:   chat.CreatorId,
			CreatedAt:   chat.CreatedAt,
			UnreadCount: unreadCounts[chat.Id],
		}

		if lastMsg, ok := lastMessages[chat.Id]; ok {
			dto.LastMessage = lastMsg.Text
			dto.LastMessageSender = lastMsg.FromUsername
			dto.LastMessageAt = lastMsg.CreatedAt
		}

		allChats = append(allChats, dto)

	}
	return allChats, nil
}
