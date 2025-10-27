package user

import (
	"context"

	"github.com/GolZrd/micro-chat/auth/internal/converter"
	descUser "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) Get(ctx context.Context, req *descUser.GetRequest) (*descUser.GetResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userObj, err := s.userService.Get(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &descUser.GetResponse{
		User: converter.ToUserFromService(userObj),
	}, nil
}
