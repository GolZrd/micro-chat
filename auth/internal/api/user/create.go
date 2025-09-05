package user

import (
	userService "auth/internal/service/user"
	descUser "auth/pkg/user_v1"
	"context"
	"log"
)

func (s *Implementation) Create(ctx context.Context, req *descUser.CreateRequest) (*descUser.CreateRespone, error) {
	// proto → service DTO
	input := userService.CreateUserDTO{
		Name:            req.Info.Name,
		Email:           req.Info.Email,
		Password:        req.Info.Password,
		PasswordConfirm: req.Info.PasswordConfirm,
		Role:            req.Info.Role.String(),
	}

	id, err := s.userService.Create(ctx, input)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return nil, err
	}

	log.Printf("Create user with id: %d", id)

	return &descUser.CreateRespone{
		Id: id,
	}, nil
}
