package user

import (
	"context"

	descUser "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) GetUsers(ctx context.Context, req *descUser.GetUsersRequest) (*descUser.GetUsersResponse, error) {
	if len(req.Ids) <= 0 {
		return nil, status.Error(codes.InvalidArgument, "ids is required")
	}

	// Ограничиваем максимальное количество пользователей
	ids := req.Ids
	if len(ids) > 100 {
		ids = ids[:100]
	}

	users, err := s.userService.GetUsers(ctx, ids)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := make([]*descUser.UserInfoShort, 0, len(users))
	for _, user := range users {
		res = append(res, &descUser.UserInfoShort{
			Id:        user.Id,
			Username:  user.Username,
			AvatarUrl: user.AvatarURL,
		})
	}

	return &descUser.GetUsersResponse{
		Users: res,
	}, nil
}
