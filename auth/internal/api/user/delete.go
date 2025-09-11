package user

import (
	"context"
	"log"

	descUser "github.com/GolZrd/micro-chat/auth/pkg/user_v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) Delete(ctx context.Context, req *descUser.DeleteRequest) (*emptypb.Empty, error) {
	err := s.userService.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	log.Printf("Delete user with id: %d", req.Id)

	return &emptypb.Empty{}, nil
}
