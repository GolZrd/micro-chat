package api

import (
	"context"

	"github.com/GolZrd/micro-chat/chat-server/internal/utils"
	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MyChats возвращает список чатов пользователя
func (s *Implementation) MyChats(ctx context.Context, req *desc.MyChatsRequest) (*desc.MyChatsResponse, error) {
	// Извлекаем userId из токена
	userId, err := utils.GetUIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication is required")
	}

	chats, err := s.chatService.MyChats(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Преобразуем в protobuf формат
	chatsPb := &desc.MyChatsResponse{
		Chats: make([]*desc.ChatInfo, 0, len(chats)),
	}

	for _, chat := range chats {
		chatsInfo := &desc.ChatInfo{
			Id:                chat.ID,
			Name:              chat.Name,
			Usernames:         chat.Usernames,
			IsDirect:          chat.IsDirect,
			IsPublic:          chat.IsPublic,
			CreatorId:         chat.CreatorId,
			CreatedAt:         timestamppb.New(chat.CreatedAt),
			LastMessage:       chat.LastMessage,
			LastMessageSender: chat.LastMessageSender,
			UnreadCount:       chat.UnreadCount,
			MemberIds:         chat.MemberIds,
		}

		if !chat.LastMessageAt.IsZero() {
			chatsInfo.LastMessageAt = timestamppb.New(chat.LastMessageAt)
		}

		chatsPb.Chats = append(chatsPb.Chats, chatsInfo)
	}

	return chatsPb, nil
}
