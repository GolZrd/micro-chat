package user

import (
	"context"

	descUser "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) SearchUser(ctx context.Context, req *descUser.SearchUserRequest) (*descUser.SearchUserResponse, error) {

	// Достаем userId из контекста
	userId, err := s.getUIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	limit := int(req.Limit)
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	users, err := s.userService.SearchUser(ctx, req.Query, userId, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	responseUsers := make([]*descUser.UserSearchResult, 0, len(users))

	for _, user := range users {
		responseUsers = append(responseUsers, &descUser.UserSearchResult{
			Id:               user.Id,
			Username:         user.Username,
			FriendshipStatus: user.FriendshipStatus,
		})
	}

	return &descUser.SearchUserResponse{
		Users: responseUsers,
	}, nil
}
