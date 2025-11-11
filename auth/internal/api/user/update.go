package user

import (
	"context"

	userService "github.com/GolZrd/micro-chat/auth/internal/service/user"
	descUser "github.com/GolZrd/micro-chat/auth/pkg/user_v1"

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
	input := userService.UpdateUserDTO{}

	if req.Info.Name != nil {
		input.Name = &req.Info.Name.Value
	}

	if req.Info.Email != nil {
		input.Email = &req.Info.Email.Value
	}

	err := s.userService.Update(ctx, req.Id, input)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
