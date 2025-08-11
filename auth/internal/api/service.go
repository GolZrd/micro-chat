package api

import (
	"auth/internal/service"
	desc "auth/pkg/auth_v1"
)

type Implementation struct {
	desc.UnimplementedAuthServer
	authService service.AuthService
}

func NewImplementation(authService service.AuthService) *Implementation {
	return &Implementation{authService: authService}
}
