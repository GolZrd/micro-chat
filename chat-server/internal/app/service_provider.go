package app

import (
	"context"
	"log"

	"github.com/GolZrd/micro-chat/chat-server/internal/api"
	"github.com/GolZrd/micro-chat/chat-server/internal/client/grpc/access"
	"github.com/GolZrd/micro-chat/chat-server/internal/closer"
	"github.com/GolZrd/micro-chat/chat-server/internal/config"
	"github.com/GolZrd/micro-chat/chat-server/internal/interceptor"
	"github.com/GolZrd/micro-chat/chat-server/internal/repository"
	"github.com/GolZrd/micro-chat/chat-server/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Описываем структуру для хранения зависимостей
type serviceProvider struct {
	cfg    *config.Config
	pgPool *pgxpool.Pool

	accessClient    *access.Client
	authInterceptor *interceptor.AuthInterceptor

	chatRepository repository.ChatRepository
	chatService    service.ChatService
	chatImpl       *api.Implementation
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

func (s *serviceProvider) AccessClient() *access.Client {
	if s.accessClient == nil {
		client, err := access.NewClient("localhost:" + s.cfg.GRPCAuthPort)
		if err != nil {
			log.Fatalf("failed to create access client: %v", err)
		}
		s.accessClient = client
	}

	return s.accessClient
}

func (s *serviceProvider) AuthInterceptor() *interceptor.AuthInterceptor {
	if s.authInterceptor == nil {
		s.authInterceptor = interceptor.NewAuthInterceptor(s.AccessClient())
	}

	return s.authInterceptor
}

func (s *serviceProvider) ChatRepository(ctx context.Context) repository.ChatRepository {
	if s.chatRepository == nil {
		s.chatRepository = repository.NewRepository(s.PgPool(ctx))
	}

	return s.chatRepository
}

func (s *serviceProvider) ChatService(ctx context.Context) service.ChatService {
	if s.chatService == nil {
		s.chatService = service.NewService(s.ChatRepository(ctx))
	}

	return s.chatService
}

func (s *serviceProvider) ChatImpl(ctx context.Context) *api.Implementation {
	if s.chatImpl == nil {
		s.chatImpl = api.NewImplementation(s.ChatService(ctx))
	}

	return s.chatImpl
}
