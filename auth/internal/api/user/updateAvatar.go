package user

import (
	"context"

	descUser "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) UpdateAvatar(ctx context.Context, req *descUser.UpdateAvatarRequest) (*descUser.UpdateAvatarResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if req.AvatarUrl == "" {
		return nil, status.Error(codes.InvalidArgument, "avatar_url is required")
	}

	err := s.userService.UpdateAvatar(ctx, req.Id, req.AvatarUrl)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &descUser.UpdateAvatarResponse{AvatarUrl: req.AvatarUrl}, nil
}
