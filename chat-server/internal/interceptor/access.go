package interceptor

import (
	"context"
	"strings"

	"github.com/GolZrd/micro-chat/chat-server/internal/client/grpc/access"
	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	serverAddress = "auth:50051"
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
			logger.Debug("no metadata in context",
				zap.String("method", info.FullMethod),
			)
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}
		authHeader, ok := md["authorization"]
		if !ok || len(authHeader) == 0 {
			logger.Debug("no authorization header",
				zap.String("method", info.FullMethod),
			)
			return nil, status.Error(codes.Unauthenticated, "authorization header is not provided")
		}

		// Проверяем есть ли у этого authHeader[0] префикс Bearer. Мы проверяем, что пришел именно JWT токен.
		if !strings.HasPrefix(authHeader[0], authPrefix) {
			logger.Debug("invalid authorization header format",
				zap.String("method", info.FullMethod),
			)
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		// Создаем новый контекст с токеном для передачи в auth сервис
		outgoingCtx := metadata.AppendToOutgoingContext(ctx, "authorization", authHeader[0])

		// Вызываем метод для проревки доступа
		err = i.accessClient.CheckAccess(outgoingCtx, info.FullMethod)
		if err != nil {
			logger.Debug("access check failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}

		// Если проверка прошла успешно, то вызываем обработчик
		return handler(ctx, req)
	}
}
