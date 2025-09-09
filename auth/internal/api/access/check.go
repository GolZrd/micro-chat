package access

import (
	descAccess "auth/pkg/access_v1"
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	authPrefix = "Bearer "
)

func (s *Implementation) Check(ctx context.Context, req *descAccess.CheckRequest) (*emptypb.Empty, error) {
	// Достаем acessToken из контекста
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	authheader, ok := md["authorization"]
	if !ok || len(authheader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization header is not provided")
	}

	// Проверяем есть ли у этого authheader[0] префикс Bearer. Мы проверяем, что пришел именно JWT токен.
	if !strings.HasPrefix(authheader[0], authPrefix) {
		return nil, status.Error(codes.InvalidArgument, "invalid authorization header format")
	}

	// Если всё ок, то отрезаем этот префикс
	accessToken := strings.TrimPrefix(authheader[0], authPrefix)

	// Получаем endpoint
	endPoint := req.GetEndpointAddress()
	if endPoint == "" {
		return nil, status.Error(codes.InvalidArgument, "endpoint_address is required")
	}

	// Вызываем сервисный слой
	err := s.accessService.Check(ctx, accessToken, endPoint)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
