package api

import (
	"context"
	"log"

	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"

	"github.com/GolZrd/micro-chat/chat-server/internal/service"
	"github.com/GolZrd/micro-chat/chat-server/internal/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	// Извлекаем username из токена
	username, err := utils.GetUsernameFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get username from token: %v", err)
	}
	// proto → service DTO
	msg := service.SendMessageDTO{
		Chat_id:       req.ChatId,
		From_username: username,
		Text:          req.Text,
		Created_at:    req.CreatedAt.AsTime(),
	}

	err = s.chatService.SendMessage(ctx, msg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send message: %v", err)
	}

	log.Printf("Send message - %v , from - %v, to chat - %v in time: %v", req.Text, req.From, req.ChatId, req.CreatedAt)

	return &emptypb.Empty{}, nil
}
