package main

import (
	"auth/internal/config"
	"auth/internal/repository"
	desc "auth/pkg/auth_v1"
	"context"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	desc.UnimplementedAuthServer
	AuthRepository repository.AuthRepository
}

// Опишем холостую логику наших ручек

// Create - ручка создания нового пользователя в системе.
func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateRespone, error) {
	id, err := s.AuthRepository.Create(ctx, req.Info)
	if err != nil {
		return nil, err
	}

	// Выведем в консоль данные переданные в запросе
	log.Printf("Inserted note with id: %d", id)

	return &desc.CreateRespone{
		Id: id,
	}, nil
}

// Get - ручка получения информации о пользователе по его идентификатору.
func (s *server) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	userObj, err := s.AuthRepository.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	log.Printf("Id: %d, Name: %s, Email: %s, Password: %s, Role: %s", userObj.Id, userObj.Info.Name, userObj.Info.Email, userObj.Info.Password, userObj.Info.Role)

	return &desc.GetResponse{
		User: userObj,
	}, nil
}

// Update - ручка обновления информации о пользователе по его идентификатору
func (s *server) Update(ctx context.Context, req *desc.UpdateRequest) (*emptypb.Empty, error) {
	err := s.AuthRepository.Update(ctx, req.Id, req.Info)
	if err != nil {
		return nil, err
	}

	// Выведем в консоль данные переданные в запросе
	log.Printf("Update user with id: %d, name: %s, email: %s", req.Id, req.Info.Name.Value, req.Info.Email.Value)

	return &emptypb.Empty{}, nil
}

// Delete - удаление пользователя из системы по его идентификатору.
func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	err := s.AuthRepository.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Выведем в консоль данные переданные в запросе
	log.Printf("Delete user with id: %d", req.Id)

	return &emptypb.Empty{}, nil
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

	s := grpc.NewServer()
	reflection.Register(s)

	desc.RegisterAuthServer(s, &server{AuthRepository: authRepo})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
