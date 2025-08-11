package api

import (
	"auth/internal/converter"
	desc "auth/pkg/auth_v1"
	"context"
	"log"
)

func (s *Implementation) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	userObj, err := s.authService.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	log.Printf("Get user with Id: %d, Name: %s, Email: %s, Password: %s, Role: %s", userObj.Id, userObj.Info.Name, userObj.Info.Email, userObj.Info.Password, userObj.Info.Role)

	return &desc.GetResponse{
		User: converter.ToUserFromService(userObj),
	}, nil
}
