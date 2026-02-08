package friends

import (
	"context"

	desc "github.com/GolZrd/micro-chat/auth/pkg/friends_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) SendFriendRequest(ctx context.Context, req *desc.SendFriendRequestRequest) (*emptypb.Empty, error) {
	if req.UserId <= 0 && req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "user id or username is required")
	}

	// Получаем id пользователя
	userId, err := s.getUIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.friendsService.SendFriendRequest(ctx, userId, req.Username, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
