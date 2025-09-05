package auth

import (
	descAuth "auth/pkg/auth_v1"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) AccessToken(ctx context.Context, req *descAuth.GetAccessTokenRequest) (*descAuth.GetAccessTokenResponse, error) {
	accessToken, err := s.authService.AccessToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &descAuth.GetAccessTokenResponse{AccessToken: accessToken}, nil
}
