package user

import (
	userService "github.com/GolZrd/micro-chat/auth/internal/service/user"
	descUser "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
)

type Implementation struct {
	descUser.UnimplementedUserAPIServer
	userService userService.UserService
}

func NewImplementation(userService userService.UserService) *Implementation {
	return &Implementation{userService: userService}
}
