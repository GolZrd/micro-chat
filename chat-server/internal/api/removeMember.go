package api

import (
	"context"

	"github.com/GolZrd/micro-chat/chat-server/internal/utils"
	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) RemoveMember(ctx context.Context, req *desc.RemoveMemberRequest) (*emptypb.Empty, error) {
	userId, err := utils.GetUIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication is required")
	}

	if req.ChatId == 0 {
		return nil, status.Error(codes.InvalidArgument, "chat id is required")
	}

	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	err = s.chatService.RemoveMember(ctx, req.ChatId, userId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove member: %v", err)
	}

	return &emptypb.Empty{}, nil
}
