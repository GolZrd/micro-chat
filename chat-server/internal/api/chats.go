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
	// Извлекаем username из токена
	username, err := utils.GetUsernameFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication is required")
	}

	if username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	chats, err := s.chatService.MyChats(ctx, username)
	if err != nil {
		// Обработать ошибку
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Преобразуем в protobuf формат
	chatsPb := &desc.MyChatsResponse{
		Chats: make([]*desc.ChatInfo, 0, len(chats)),
	}

	for _, chat := range chats {
		chatsInfo := &desc.ChatInfo{
			Id:        chat.ID,
			Usernames: chat.Usernames,
			CreatedAt: timestamppb.New(chat.CreatedAt),
		}
		chatsPb.Chats = append(chatsPb.Chats, chatsInfo)
	}

	return chatsPb, nil
}
