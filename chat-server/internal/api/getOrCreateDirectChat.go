package api

import (
	"context"

	"github.com/GolZrd/micro-chat/chat-server/internal/utils"
	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) GetOrCreateDirectChat(ctx context.Context, req *desc.GetOrCreateDirectChatRequest) (*desc.GetOrCreateDirectChatResponse, error) {
	// Достаем из токена username и uid пользователя
	user, err := utils.GetUserClaimsFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user claims from token: %v", err)
	}

	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	chatId, created, err := s.chatService.GetOrCreateDirectChat(ctx, user.UID, user.Username, req.UserId, req.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get or create direct chat: %v", err)
	}

	return &desc.GetOrCreateDirectChatResponse{ChatId: chatId, Created: created}, nil
}
