package app

import (
	userAPI "auth/internal/api/user"
	"auth/internal/closer"
	"auth/internal/config"
	userRepository "auth/internal/repository/user"
	userService "auth/internal/service/user"
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Описываем структуру для хранения зависимостей
type serviceProvider struct {
	cfg    *config.Config
	pgPool *pgxpool.Pool

	authRepository userRepository.AuthRepository
	authService    userService.AuthService
	authImpl       *userAPI.Implementation
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

func (s *serviceProvider) AuthRepository(ctx context.Context) userRepository.AuthRepository {
	if s.authRepository == nil {
		s.authRepository = userRepository.NewRepository(s.PgPool(ctx))
	}

	return s.authRepository
}

func (s *serviceProvider) AuthService(ctx context.Context) userService.AuthService {
	if s.authService == nil {
		s.authService = userService.NewService(s.AuthRepository(ctx))
	}

	return s.authService
}

func (s *serviceProvider) AuthImpl(ctx context.Context) *userAPI.Implementation {
	if s.authImpl == nil {
		s.authImpl = userAPI.NewImplementation(s.AuthService(ctx))
	}

	return s.authImpl
}
