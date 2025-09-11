package api

import (
	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"

	"github.com/GolZrd/micro-chat/chat-server/internal/service"
)

type Implementation struct {
	desc.UnimplementedChatServer
	chatService service.ChatService
}

func NewImplementation(chatService service.ChatService) *Implementation {
	return &Implementation{chatService: chatService}
}
