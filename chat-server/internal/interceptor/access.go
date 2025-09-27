package interceptor

import (
	"context"
	"strings"

	"github.com/GolZrd/micro-chat/chat-server/internal/client/grpc/access"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	serverAddress = "localhost:50051"
	authPrefix    = "Bearer "
)

type AuthInterceptor struct {
	accessClient *access.Client
}

func NewAuthInterceptor(accessClient *access.Client) *AuthInterceptor {
	return &AuthInterceptor{
		accessClient: accessClient,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// Извлекаем токен из метадаты
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}
		authHeader, ok := md["authorization"]
		if !ok || len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization header is not provided")
		}

		// Проверяем есть ли у этого authHeader[0] префикс Bearer. Мы проверяем, что пришел именно JWT токен.
		if !strings.HasPrefix(authHeader[0], authPrefix) {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		// Создаем новый контекст с токеном для передачи в auth сервис
		outgoingCtx := metadata.AppendToOutgoingContext(ctx, "authorization", authHeader[0])

		// Вызываем метод для проревки доступа
		err = i.accessClient.CheckAccess(outgoingCtx, info.FullMethod)
		if err != nil {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}

		// Если проверка прошла успешно, то вызываем обработчик
		return handler(ctx, req)
	}
}
