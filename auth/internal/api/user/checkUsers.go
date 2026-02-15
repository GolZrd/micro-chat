package user

import (
	"context"

	descUser "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) CheckUsersExists(ctx context.Context, req *descUser.CheckUsersExistsRequest) (*descUser.CheckUsersExistsResponse, error) {
	if len(req.Usernames) == 0 {
		return nil, status.Error(codes.InvalidArgument, "usernames is required")
	}

	foundUsers, notFoundUsers, err := s.userService.CheckUsersExists(ctx, req.Usernames)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var foundProto []*descUser.UserInfoShort
	for _, user := range foundUsers {
		foundProto = append(foundProto, &descUser.UserInfoShort{
			Id:       user.Id,
			Username: user.Username,
		})
	}

	return &descUser.CheckUsersExistsResponse{NotFoundUsers: notFoundUsers, FoundUsers: foundProto}, nil
}
