package auth

import (
	AuthService "github.com/GolZrd/micro-chat/auth/internal/service/auth"
	descAuth "github.com/GolZrd/micro-chat/auth/pkg/auth_v1"
)

type Implementation struct {
	descAuth.UnimplementedAuthAPIServer
	authService AuthService.AuthService
}

func NewImplementation(authService AuthService.AuthService) *Implementation {
	return &Implementation{authService: authService}
}
