package api

import (
	"context"

	"github.com/GolZrd/micro-chat/chat-server/internal/utils"
	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) MarkChatRead(ctx context.Context, req *desc.MarkChatReadRequest) (*emptypb.Empty, error) {
	if req.ChatId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "chat id is required")
	}

	userId, err := utils.GetUIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication is required")
	}

	err = s.chatService.MarkChatRead(ctx, req.ChatId, userId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark chat as read: %v", err)
	}

	return &emptypb.Empty{}, nil
}
