package user

import (
	"context"

	userService "github.com/GolZrd/micro-chat/auth/internal/service/user"
	descUser "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) Create(ctx context.Context, req *descUser.CreateRequest) (*descUser.CreateRespone, error) {
	// Валидируем входные данные
	if err := validateCreateUser(req); err != nil {
		return nil, err
	}

	// proto → service DTO
	input := userService.CreateUserDTO{
		Username:        req.Info.Username,
		Email:           req.Info.Email,
		Password:        req.Info.Password,
		PasswordConfirm: req.Info.PasswordConfirm,
		Role:            req.Info.Role.String(),
	}

	id, err := s.userService.Create(ctx, input)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &descUser.CreateRespone{
		Id: id,
	}, nil
}

func validateCreateUser(req *descUser.CreateRequest) error {
	if req.Info == nil {
		return status.Error(codes.InvalidArgument, "info is required")
	}

	if req.Info.Username == "" {
		return status.Error(codes.InvalidArgument, "name is required")
	}

	if req.Info.Email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Info.Password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.Info.PasswordConfirm == "" {
		return status.Error(codes.InvalidArgument, "password_confirm is required")
	}

	if req.Info.Role.String() == "" {
		return status.Error(codes.InvalidArgument, "role is required")
	}

	return nil
}
