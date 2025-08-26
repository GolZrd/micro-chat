package user

import (
	userService "auth/internal/service/user"
	descUser "auth/pkg/user_v1"
)

type Implementation struct {
	descUser.UnimplementedAuthServer
	authService userService.AuthService
}

func NewImplementation(authService userService.AuthService) *Implementation {
	return &Implementation{authService: authService}
}
