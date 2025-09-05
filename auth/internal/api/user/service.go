package user

import (
	userService "auth/internal/service/user"
	descUser "auth/pkg/user_v1"
)

type Implementation struct {
	descUser.UnimplementedAuthServer
	userService userService.UserService
}

func NewImplementation(userService userService.UserService) *Implementation {
	return &Implementation{userService: userService}
}
