package api

import (
	"context"

	"github.com/GolZrd/micro-chat/chat-server/internal/utils"
	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) Heartbeat(ctx context.Context, req *desc.HeartbeatRequest) (*emptypb.Empty, error) {
	// Получаем userId из контекста
	userId, err := utils.GetUIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication is required")
	}

	// Вызываем метод сервиса
	err = s.chatService.Heartbeat(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
