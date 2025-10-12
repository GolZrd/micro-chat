package api

import (
	"context"
	"log"

	"github.com/GolZrd/micro-chat/chat-server/internal/utils"
	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	// Достаем из токена username пользователя
	creatorUsername, err := utils.GetUsernameFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get username from token: %v", err)
	}

	log.Printf("CreateChat: creator=%s, inviting=%v", creatorUsername, req.Usernames)

	// Создаем срез участников
	participants := make([]string, 0, len(req.Usernames)+1)

	// Добавляем создателя
	participants = append(participants, creatorUsername)

	// Добавляем остальных участников чата, и проверяем чтобы не было создателя
	for _, username := range req.Usernames {
		if username == "" || username == creatorUsername {
			continue
		}

		participants = append(participants, username)
	}

	id, err := s.chatService.Create(ctx, participants)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create chat: %v", err)
	}

	log.Printf("Created chat with id: %d", id)

	return &desc.CreateResponse{
		ChatId: id,
	}, nil

}
