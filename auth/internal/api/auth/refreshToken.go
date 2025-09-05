package auth

import (
	descAuth "auth/pkg/auth_v1"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) RefreshToken(ctx context.Context, req *descAuth.GetRefreshTokenRequest) (*descAuth.GetRefreshTokenResponse, error) {

	refreshToken, err := s.authService.RefreshToken(ctx, req.GetOldRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &descAuth.GetRefreshTokenResponse{RefreshToken: refreshToken}, nil
}
