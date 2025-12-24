package interceptor

import (
	"context"
	"strings"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func MetricsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()
	method := extractMethod(info.FullMethod)

	// Увеличиваем счетчик активных запросов
	metric.IncRequestsInFlight()
	defer metric.DecRequestsInFlight()

	resp, err = handler(ctx, req)

	status := extractStatus(err)

	// Увеличиваем счетчик общего количества запросов
	metric.IncRequestTotal(status, method)

	metric.ObserveRequestDuration(method, time.Since(start).Seconds())

	return resp, err
}

// extractMethod извлекает название метода из полного пути
func extractMethod(fullMethod string) string {
	parts := strings.Split(fullMethod, "/")

	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return fullMethod
}

func extractStatus(err error) string {
	if err != nil {
		if s, ok := status.FromError(err); ok {
			return s.Code().String()
		}
	}

	return "OK"
}
