package friends

import (
	"context"

	desc "github.com/GolZrd/micro-chat/auth/pkg/friends_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) RejectFriendRequest(ctx context.Context, req *desc.RejectFriendRequestRequest) (*emptypb.Empty, error) {
	if req.RequestId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "request id is required")
	}

	// Получаем id пользователя
	userId, err := s.getUIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.friendsService.RejectFriendRequest(ctx, req.RequestId, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
