package api

import (
	"chat-server/internal/service"
	desc "chat-server/pkg/chat_v1"
	"context"
	"log"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	// proto → service DTO
	msg := service.SendMessageDTO{
		Chat_id:       req.ChatId,
		From_username: req.From,
		Text:          req.Text,
		Created_at:    req.CreatedAt.AsTime(),
	}

	err := s.chatService.SendMessage(ctx, msg)
	if err != nil {
		return nil, err
	}

	log.Printf("Send message - %v , from - %v, to chat - %v in time: %v", req.Text, req.From, req.ChatId, req.CreatedAt)

	return &emptypb.Empty{}, nil
}
