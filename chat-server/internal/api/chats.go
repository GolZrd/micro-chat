package api

import (
	"context"
	"log"

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
		return nil, status.Errorf(codes.Unauthenticated, "failed to get username from token: %v", err)
	}

	// Проверка
	log.Printf("Loading chats for user: %s", username)

	if username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	chats, err := s.chatService.MyChats(ctx, username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user chats: %v", err)
	}

	// Преобразуем в protobuf формат
	chatsPb := &desc.MyChatsResponse{
		Chats: make([]*desc.ChatInfo, len(chats)),
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
