package api

import (
	"auth/internal/service"
	desc "auth/pkg/auth_v1"
	"context"
	"log"
)

func (s *Implementation) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateRespone, error) {
	// proto → service DTO
	input := service.CreateUserDTO{
		Name:            req.Info.Name,
		Email:           req.Info.Email,
		Password:        req.Info.Password,
		PasswordConfirm: req.Info.PasswordConfirm,
		Role:            req.Info.Role.String(),
	}

	id, err := s.authService.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	log.Printf("Create user with id: %d", id)

	return &desc.CreateRespone{
		Id: id,
	}, nil
}
