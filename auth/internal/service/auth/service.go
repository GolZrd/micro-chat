package auth

import (
	"context"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/config"
	"github.com/GolZrd/micro-chat/auth/internal/model"
	authRepository "github.com/GolZrd/micro-chat/auth/internal/repository/auth"
)

type AuthService interface {
	Login(ctx context.Context, email string, password string) (refreshToken string, userId int64, err error)
	RefreshToken(ctx context.Context, oldRefreshToken string) (refreshToken string, err error)
	AccessToken(ctx context.Context, refreshToken string) (accessToken string, err error)
}

// Нам нужно получить информацию о пользователе по email, для этого нужно обратиться к репозиторию user
// Чтобы не тянуть весь userRepository.Repository мы объявим узкий интерфейс с методом GetByEmail
type UsersReader interface {
	GetByEmail(ctx context.Context, email string) (*model.UserAuthData, error)
}

type service struct {
	authRepository   authRepository.AuthRepository
	userRepository   UsersReader
	RefreshSecretKey string
	AccessSecretKey  string
	accessTTL        time.Duration
	refreshTTL       time.Duration
}

func NewService(authRepository authRepository.AuthRepository, users UsersReader, cfg *config.Config) AuthService {
	return &service{authRepository: authRepository,
		userRepository:   users,
		RefreshSecretKey: cfg.RefreshSecretKey,
		AccessSecretKey:  cfg.AccessSecretKey,
		accessTTL:        cfg.AccessTTL,
		refreshTTL:       cfg.RefreshTTL}
}
