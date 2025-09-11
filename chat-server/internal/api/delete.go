package api

import (
	"context"
	"log"

	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {

	err := s.chatService.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	log.Printf("Delete chat with id: %d", req.Id)

	return &emptypb.Empty{}, nil
}
