package main

import (
	"chat-server/internal/config"
	"chat-server/internal/repository"
	"chat-server/internal/repository/model"
	desc "chat-server/pkg/chat_v1"
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
	desc.UnimplementedChatServer
	ChatRepository repository.ChatRepository
}

// Create - ручка создания нового чата.
func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	log.Printf("Create chat with usernames: %v", req.Usernames)
	id, err := s.ChatRepository.Create(ctx, req.GetUsernames())
	if err != nil {
		log.Printf("Failed to create chat: %v", err)
		return nil, err
	}

	log.Printf("Inserted chat with id: %d", id)

	return &desc.CreateResponse{
		ChatId: id,
	}, nil
}

// Delete - удаление чата из системы по его идентификатору.
func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	log.Printf("Delete chat with id: %d", req.GetId())

	err := s.ChatRepository.Delete(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// SendMessage - ручка отправки сообщения на сервер.
func (s *server) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	log.Printf("Send message - %v , from - %v, to chat - %v in time: %v", req.Text, req.From, req.ChatId, req.CreatedAt)

	// Используем модель message
	err := s.ChatRepository.SendMessage(ctx, model.Message{
		ChatId:       req.GetChatId(),
		Text:         req.GetText(),
		FromUsername: req.GetFrom(),
		CreatedAt:    req.GetCreatedAt().AsTime(),
	})
	if err != nil {
		return nil, err
	}

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

	log.Printf("config: %v", cfg)
	// Создаем пул соединений с БД
	pool, err := pgxpool.New(ctx, cfg.DB_DSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	chatRepo := repository.NewRepository(pool)

	s := grpc.NewServer()
	reflection.Register(s)

	desc.RegisterChatServer(s, &server{ChatRepository: chatRepo})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
