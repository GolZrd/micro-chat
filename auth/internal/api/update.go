package api

import (
	"auth/internal/service"
	desc "auth/pkg/auth_v1"
	"context"
	"log"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Implementation) Update(ctx context.Context, req *desc.UpdateRequest) (*emptypb.Empty, error) {
	// proto → service DTO
	input := service.UpdateUserDTO{
		Name:  &req.Info.Name.Value,
		Email: &req.Info.Email.Value,
	}

	err := s.authService.Update(ctx, req.Id, input)
	if err != nil {
		return nil, err
	}

	log.Printf("Update user with id: %d, name: %s, email: %s", req.Id, req.Info.Name.Value, req.Info.Email.Value)

	return &emptypb.Empty{}, nil
}
