package friends

import (
	"context"

	desc "github.com/GolZrd/micro-chat/auth/pkg/friends_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) GetFriends(ctx context.Context, req *desc.GetFriendsRequest) (*desc.GetFriendsResponse, error) {
	// Получаем id пользователя
	userId, err := s.getUIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	friendsList, err := s.friendsService.Friends(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := make([]*desc.Friend, 0, len(friendsList))

	for _, friend := range friendsList {
		result = append(result, &desc.Friend{
			Id:       friend.Id,
			UserId:   friend.UserId,
			Username: friend.Username,
		})
	}

	return &desc.GetFriendsResponse{
		Friends: result,
	}, nil
}
