package interceptor

import (
	"context"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LogInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// Логируем начало запроса
	logger.Info("request started",
		zap.String("method", info.FullMethod),
	)

	start := time.Now()

	// Вызываем handler
	resp, err = handler(ctx, req)

	// Логируем результат
	duration := time.Since(start)

	if err != nil {
		code := status.Code(err)

		// Клиентские ошибки - не ERROR
		if code == codes.InvalidArgument || code == codes.NotFound ||
			code == codes.Unauthenticated || code == codes.PermissionDenied {
			logger.Warn("client error",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
				zap.String("code", code.String()),
			)
		} else {
			logger.Error("server error",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		}
	} else {
		logger.Info("request completed",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
		)
	}

	return resp, err
}
