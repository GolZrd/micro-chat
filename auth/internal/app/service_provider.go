package app

import (
	authAPI "auth/internal/api/auth"
	userAPI "auth/internal/api/user"
	"auth/internal/closer"
	"auth/internal/config"
	authRepository "auth/internal/repository/auth"
	userRepository "auth/internal/repository/user"
	authService "auth/internal/service/auth"
	userService "auth/internal/service/user"
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Описываем структуру для хранения зависимостей
type serviceProvider struct {
	cfg    *config.Config
	pgPool *pgxpool.Pool

	userRepository userRepository.UserRepository
	authRepository authRepository.AuthRepository
	userService    userService.UserService
	authService    authService.AuthService
	userImpl       *userAPI.Implementation
	authImpl       *authAPI.Implementation
}

func newServiceProvider() *serviceProvider {
	return &serviceProvider{}
}

// Функция, которая проверяет был ли уже инициализирован конфиг. Если нет, то инициализируем его. и возвращаем.
func (s *serviceProvider) Config() *config.Config {
	if s.cfg == nil {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		s.cfg = cfg
	}

	return s.cfg
}

func (s *serviceProvider) PgPool(ctx context.Context) *pgxpool.Pool {
	if s.pgPool == nil {
		pool, err := pgxpool.New(ctx, s.Config().DB_DSN)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}

		err = pool.Ping(ctx)
		if err != nil {
			log.Fatalf("failed to ping database: %v", err)
		}

		closer.Add(func() error {
			pool.Close()
			return nil
		})

		s.pgPool = pool
	}

	return s.pgPool
}

func (s *serviceProvider) UserRepository(ctx context.Context) userRepository.UserRepository {
	if s.userRepository == nil {
		s.userRepository = userRepository.NewRepository(s.PgPool(ctx))
	}

	return s.userRepository
}

func (s *serviceProvider) UserService(ctx context.Context) userService.UserService {
	if s.userService == nil {
		s.userService = userService.NewService(s.UserRepository(ctx))
	}

	return s.userService
}

func (s *serviceProvider) UserImpl(ctx context.Context) *userAPI.Implementation {
	if s.userImpl == nil {
		s.userImpl = userAPI.NewImplementation(s.UserService(ctx))
	}

	return s.userImpl
}

func (s *serviceProvider) AuthRepository(ctx context.Context) authRepository.AuthRepository {
	if s.authRepository == nil {
		s.authRepository = authRepository.NewRepository(s.PgPool(ctx))
	}

	return s.authRepository
}

func (s *serviceProvider) AuthService(ctx context.Context) authService.AuthService {
	if s.authService == nil {
		s.authService = authService.NewService(s.AuthRepository(ctx), s.UserRepository(ctx), s.Config())
	}

	return s.authService
}

func (s *serviceProvider) AuthImpl(ctx context.Context) *authAPI.Implementation {
	if s.authImpl == nil {
		s.authImpl = authAPI.NewImplementation(s.AuthService(ctx))
	}

	return s.authImpl
}
