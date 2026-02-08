package friends

import (
	"context"
	"strings"

	"github.com/GolZrd/micro-chat/auth/internal/config"
	friendsService "github.com/GolZrd/micro-chat/auth/internal/service/friends"
	"github.com/GolZrd/micro-chat/auth/internal/utils/jwt"
	descFriends "github.com/GolZrd/micro-chat/auth/pkg/friends_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Implementation struct {
	descFriends.UnimplementedFriendsAPIServer
	friendsService friendsService.FriendsService

	jwtManager jwt.JWTManager
	secretKey  string
}

func NewImplementation(friendsService friendsService.FriendsService, jwtManager jwt.JWTManager, cfg *config.Config) *Implementation {
	return &Implementation{friendsService: friendsService, jwtManager: jwtManager, secretKey: cfg.AccessSecretKey}
}

// Вспомогательный метод для получения UID из контекста
func (i *Implementation) getUIDFromContext(ctx context.Context) (int64, error) {
	// Извлекаем токен из метадаты
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return 0, status.Error(codes.Unauthenticated, "authorization header is not provided")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")

	// Проверяем токен
	claims, err := i.jwtManager.VerifyToken(token, []byte(i.secretKey))
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "invalid token")
	}

	return claims.UID, nil

}
