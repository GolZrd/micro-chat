package auth

import (
	descAuth "auth/pkg/auth_v1"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) Login(ctx context.Context, req *descAuth.LoginRequest) (*descAuth.LoginResponse, error) {
	// Валидируем данные
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	refreshToken, err := s.authService.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		//TODO: обработка в зависимости от ошибки
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &descAuth.LoginResponse{RefreshToken: refreshToken}, nil
}
