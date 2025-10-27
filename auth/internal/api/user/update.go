package user

import (
	"context"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	userService "github.com/GolZrd/micro-chat/auth/internal/service/user"
	descUser "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) Update(ctx context.Context, req *descUser.UpdateRequest) (*emptypb.Empty, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if req.Info == nil {
		return nil, status.Error(codes.InvalidArgument, "info is required")
	}

	if req.Info.Name == nil && req.Info.Email == nil {
		return nil, status.Error(codes.InvalidArgument, "at least one field is required")
	}

	// proto → service DTO
	input := userService.UpdateUserDTO{
		Name:  &req.Info.Name.Value,
		Email: &req.Info.Email.Value,
	}

	logger.Debug("updating user",
		zap.Int64("user_id", req.Id),
	)

	err := s.userService.Update(ctx, req.Id, input)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
