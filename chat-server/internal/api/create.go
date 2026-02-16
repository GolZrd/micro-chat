package api

import (
	"context"
	"errors"
	"strings"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"github.com/GolZrd/micro-chat/chat-server/internal/service"
	"github.com/GolZrd/micro-chat/chat-server/internal/utils"
	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	if len(req.Usernames) == 0 {
		return nil, status.Error(codes.InvalidArgument, "usernames is required")
	}

	// Достаем из токена username и uid пользователя
	user, err := utils.GetUserClaimsFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user claims from token: %v", err)
	}
	logger.Debug("attempt to create chat", zap.String("creator", user.Username), zap.Strings("inviting", req.Usernames))

	// Передаем создателя отдельно от приглашенных
	id, err := s.chatService.Create(ctx, req.Name, user.UID, user.Username, req.Usernames)
	if err != nil {
		// Проверяем типизированную ошибку
		var usersNotFound *service.ErrUserNotFound
		if errors.As(err, &usersNotFound) {
			return nil, status.Errorf(codes.NotFound, "USERS_NOT_FOUND:%s", strings.Join(usersNotFound.Usernames, ","))
		}
		return nil, status.Errorf(codes.Internal, "failed to create chat: %v", err)
	}

	logger.Debug("Created chat", zap.Int64("chat_id", id), zap.Strings("participants", req.Usernames))

	return &desc.CreateResponse{
		ChatId: id,
	}, nil

}
