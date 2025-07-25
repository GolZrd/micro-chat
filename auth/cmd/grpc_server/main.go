package main

import (
	desc "auth/pkg/auth_v1"
	"context"
	"fmt"
	"log"
	"net"

	"github.com/brianvoe/gofakeit/v6"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const grpcPort = 50051

type server struct {
	desc.UnimplementedAuthServer
}

// Опишем холостую логику наших ручек

// Create - ручка создания нового пользователя в системе.
func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateRespone, error) {
	// Выведем в консоль данные переданные в запросе
	log.Printf("Create user with name: %s, email: %s, role: %v", req.Name, req.Email, req.Role)

	return &desc.CreateRespone{
		Id: 0,
	}, nil
}

// Get - ручка получения информации о пользователе по его идентификатору.
func (s *server) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	// Выведем в консоль данные переданные в запросе
	log.Printf("Get user with id: %d", req.Id)

	return &desc.GetResponse{
		Id:        req.Id,
		Name:      gofakeit.Name(),
		Email:     gofakeit.Email(),
		Role:      desc.Role_guest,
		CreatedAt: timestamppb.New(gofakeit.Date()),
		UpdatedAt: timestamppb.New(gofakeit.Date()),
	}, nil
}

// Update - ручка обновления информации о пользователе по его идентификатору
func (s *server) Update(ctx context.Context, req *desc.UpdateRequest) (*emptypb.Empty, error) {
	// Выведем в консоль данные переданные в запросе
	log.Printf("Update user with id: %d, name: %s, email: %s", req.Id, req.Name.Value, req.Email.Value)

	return &emptypb.Empty{}, nil

}

// Delete - удаление пользователя из системы по его идентификатору.
func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	// Выведем в консоль данные переданные в запросе
	log.Printf("Delete user with id: %d", req.Id)

	return &emptypb.Empty{}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	desc.RegisterAuthServer(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
