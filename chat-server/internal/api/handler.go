package api

import (
	"chat-server/internal/service"
	desc "chat-server/pkg/chat_v1"
)

type Implementation struct {
	desc.UnimplementedChatServer
	chatService service.ChatService
}

func NewImplementation(chatService service.ChatService) *Implementation {
	return &Implementation{chatService: chatService}
}
