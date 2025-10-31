package api

import (
	"context"

	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "chat id is required")
	}

	err := s.chatService.Delete(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete chat: %v", err)
	}

	return &emptypb.Empty{}, nil
}
