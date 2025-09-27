package api

import (
	"context"
	"log"

	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	log.Printf("Create chat with usernames: %v", req.Usernames)

	id, err := s.chatService.Create(ctx, req.Usernames)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create chat: %v", err)
	}

	log.Printf("Created chat with id: %d", id)

	return &desc.CreateResponse{
		ChatId: id,
	}, nil

}
