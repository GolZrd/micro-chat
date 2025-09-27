package api

import (
	"context"
	"log"

	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {

	err := s.chatService.Delete(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete chat: %v", err)
	}

	log.Printf("Delete chat with id: %d", req.Id)

	return &emptypb.Empty{}, nil
}
