package api

import (
	"context"

	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) MyChats(ctx context.Context, req *desc.MyChatsRequest) (*desc.MyChatsResponse, error) {
	username := req.Username
	if username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	chats, err := s.chatService.MyChats(ctx, username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get my chats: %v", err)
	}

	return &desc.MyChatsResponse{chats}, nil
}
