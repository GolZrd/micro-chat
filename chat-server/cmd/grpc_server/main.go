package main

import (
	desc "chat-server/pkg/chat_v1"
	"context"
	"fmt"
	"log"
	"net"

	"github.com/brianvoe/gofakeit/v6"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

const grpcPort = 50052

type server struct {
	desc.UnimplementedChatServer
}

// Create - ручка создания нового чата.
func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	log.Printf("Create chat with usernames: %v", req.Usernames)

	return &desc.CreateResponse{
		ChatId: gofakeit.Int64(),
	}, nil
}

// Delete - удаление чата из системы по его идентификатору.
func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	log.Printf("Delete chat with id: %d", req.Id)

	return &emptypb.Empty{}, nil
}

// SendMessage - ручка отправки сообщения на сервер.
func (s *server) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	log.Printf("Send message %v from %v in %v", req.Text, req.From, req.CreatedAt)

	return &emptypb.Empty{}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	desc.RegisterChatServer(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
