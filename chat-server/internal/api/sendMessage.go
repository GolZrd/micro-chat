package api

import (
	"context"
	"errors"

	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"

	"github.com/GolZrd/micro-chat/chat-server/internal/service"
	"github.com/GolZrd/micro-chat/chat-server/internal/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	// Но не проверяем From, так как его мы достаем из токена
	if err := validateMessage(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Извлекаем username из токена
	username, err := utils.GetUsernameFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get username from token: %v", err)
	}
	// proto → service DTO
	msg := service.SendMessageDTO{
		ChatId:       req.ChatId,
		FromUsername: username,
		Text:         req.Text,
		CreatedAt:    req.CreatedAt.AsTime(),
	}

	err = s.chatService.SendMessage(ctx, msg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send message: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func validateMessage(msg *desc.SendMessageRequest) error {
	if msg.ChatId <= 0 {
		return errors.New("chat_id cannot be empty")
	}
	if msg.Text == "" {
		return errors.New("text cannot be empty")
	}
	return nil
}
