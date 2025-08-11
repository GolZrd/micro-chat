package api

import (
	desc "chat-server/pkg/chat_v1"
	"context"
	"log"
)

func (s *Implementation) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	log.Printf("Create chat with usernames: %v", req.Usernames)

	id, err := s.chatService.Create(ctx, req.Usernames)
	if err != nil {
		return nil, err
	}

	log.Printf("Created chat with id: %d", id)

	return &desc.CreateResponse{
		ChatId: id,
	}, nil

}
