package access

import (
	accessService "auth/internal/service/access"
	descAccess "auth/pkg/access_v1"
)

type Implementation struct {
	descAccess.UnimplementedAccessAPIServer
	accessService accessService.AccessService
}

func NewImplementation(accessService accessService.AccessService) *Implementation {
	return &Implementation{accessService: accessService}
}
