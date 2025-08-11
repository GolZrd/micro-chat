package main

import (
	"auth/internal/api"
	"auth/internal/config"
	"auth/internal/repository"
	"auth/internal/service"
	desc "auth/pkg/auth_v1"
	"context"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	desc.UnimplementedAuthServer
	AuthRepository repository.AuthRepository
}

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Создаем пул соединений с БД
	pool, err := pgxpool.New(ctx, cfg.DB_DSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	authRepo := repository.NewRepository(pool)
	authService := service.NewService(authRepo)

	s := grpc.NewServer()
	reflection.Register(s)

	desc.RegisterAuthServer(s, api.NewImplementation(authService))

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
