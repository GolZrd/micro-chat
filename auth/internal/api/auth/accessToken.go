package auth

import (
	descAuth "auth/pkg/auth_v1"
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) GetAccessToken(ctx context.Context, req *descAuth.GetAccessTokenRequest) (*descAuth.GetAccessTokenResponse, error) {
	accessToken, err := s.authService.AccessToken(ctx, req.GetRefreshToken())
	//Логи добавить
	log.Println("Get access token with refresh token: ", req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &descAuth.GetAccessTokenResponse{AccessToken: accessToken}, nil
}
