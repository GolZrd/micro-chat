package auth

import (
	AuthService "auth/internal/service/auth"
	descAuth "auth/pkg/auth_v1"
)

type Implementation struct {
	descAuth.UnimplementedAuthAPIServer
	authService AuthService.AuthService
}

func NewImplementation(authService AuthService.AuthService) *Implementation {
	return &Implementation{authService: authService}
}
