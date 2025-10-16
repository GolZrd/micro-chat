package app

import (
	"context"
	"log"

	accessAPI "github.com/GolZrd/micro-chat/auth/internal/api/access"
	authAPI "github.com/GolZrd/micro-chat/auth/internal/api/auth"
	userAPI "github.com/GolZrd/micro-chat/auth/internal/api/user"
	"github.com/GolZrd/micro-chat/auth/internal/cache"
	"github.com/GolZrd/micro-chat/auth/internal/closer"
	"github.com/GolZrd/micro-chat/auth/internal/config"
	accessRepository "github.com/GolZrd/micro-chat/auth/internal/repository/access"
	authRepository "github.com/GolZrd/micro-chat/auth/internal/repository/auth"
	userRepository "github.com/GolZrd/micro-chat/auth/internal/repository/user"
	accessService "github.com/GolZrd/micro-chat/auth/internal/service/access"
	authService "github.com/GolZrd/micro-chat/auth/internal/service/auth"
	userService "github.com/GolZrd/micro-chat/auth/internal/service/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Описываем структуру для хранения зависимостей
type serviceProvider struct {
	cfg        *config.Config
	pgPool     *pgxpool.Pool
	redisCache *cache.RedisCache

	userRepository   userRepository.UserRepository
	authRepository   authRepository.AuthRepository
	accessRepository accessRepository.AccessRepository
	userService      userService.UserService
	authService      authService.AuthService
	accessService    accessService.AccessService
	userImpl         *userAPI.Implementation
	authImpl         *authAPI.Implementation
	accessImpl       *accessAPI.Implementation
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

func (s *serviceProvider) RedisCache(ctx context.Context) *cache.RedisCache {
	if s.redisCache == nil {

		redisCache, err := cache.NewRedisCache(s.Config().RedisAddr, s.Config().RedisPassword, s.Config().RedisDB, s.Config().RedisTTL)
		if err != nil {
			log.Fatalf("failed to connect to redis: %v", err)
		}

		closer.Add(func() error {
			redisCache.Close()
			return nil
		})

		s.redisCache = redisCache
	}

	return s.redisCache
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

func (s *serviceProvider) AccessRepository(ctx context.Context) accessRepository.AccessRepository {
	if s.accessRepository == nil {
		s.accessRepository = accessRepository.NewRepository(s.PgPool(ctx))
	}

	return s.accessRepository
}

func (s *serviceProvider) AccessService(ctx context.Context) accessService.AccessService {
	if s.accessService == nil {
		s.accessService = accessService.NewService(s.AccessRepository(ctx), s.RedisCache(ctx), s.Config())
	}

	return s.accessService
}

func (s *serviceProvider) AccessImpl(ctx context.Context) *accessAPI.Implementation {
	if s.accessImpl == nil {
		s.accessImpl = accessAPI.NewImplementation(s.AccessService(ctx))
	}

	return s.accessImpl
}
