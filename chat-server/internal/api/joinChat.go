package api

import (
	"context"

	"github.com/GolZrd/micro-chat/chat-server/internal/utils"
	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) JoinChat(ctx context.Context, req *desc.JoinChatRequest) (*emptypb.Empty, error) {
	user, err := utils.GetUserClaimsFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	if req.ChatId == 0 {
		return nil, status.Error(codes.InvalidArgument, "chat id is required")
	}

	err = s.chatService.JoinChat(ctx, req.ChatId, user.UID, user.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to join chat: %v", err)
	}

	return &emptypb.Empty{}, nil
}
