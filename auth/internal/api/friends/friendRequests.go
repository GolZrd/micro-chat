package friends

import (
	"context"

	desc "github.com/GolZrd/micro-chat/auth/pkg/friends_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Implementation) GetFriendRequests(ctx context.Context, req *desc.GetFriendRequestsRequest) (*desc.GetFriendRequestsResponse, error) {
	// Получаем id пользователя
	userId, err := s.getUIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	requests, err := s.friendsService.FriendRequests(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := make([]*desc.FriendRequest, 0, len(requests))

	for _, req := range requests {
		result = append(result, &desc.FriendRequest{
			Id:           req.Id,
			FromUserId:   req.FromUserId,
			FromUsername: req.FromUsername,
			CreatedAt:    timestamppb.New(req.CreatedAt),
		})
	}

	return &desc.GetFriendRequestsResponse{
		Requests: result,
	}, nil
}
