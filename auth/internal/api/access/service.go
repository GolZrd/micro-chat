package access

import (
	accessService "github.com/GolZrd/micro-chat/auth/internal/service/access"
	descAccess "github.com/GolZrd/micro-chat/auth/pkg/access_v1"
)

type Implementation struct {
	descAccess.UnimplementedAccessAPIServer
	accessService accessService.AccessService
}

func NewImplementation(accessService accessService.AccessService) *Implementation {
	return &Implementation{accessService: accessService}
}
