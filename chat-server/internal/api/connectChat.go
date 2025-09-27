package api

import (
	"io"
	"log"

	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Implementation) ConnectChat(req *desc.ConnectChatRequest, stream desc.Chat_ConnectChatServer) error {
	log.Printf("Client connecting to chat %d", req.ChatId)

	// Подписываемся на сообщения чата
	msgChan, err := s.chatService.ConnectToChat(stream.Context(), req.ChatId)
	if err != nil {
		return status.Errorf(codes.NotFound, "failed to subscribe: %v", err)
	}

	// Горутина в SubscribeToChat автоматически отпишется при отмене контекста
	defer log.Printf("Client disconnected from chat %d", req.ChatId)

	// Слушаем канал и отправляем сообщения клиенту
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return nil
			}
			// Преобразуем в proto
			protoMsg := &desc.Message{
				From:      msg.From,
				Text:      msg.Text,
				CreatedAt: timestamppb.New(msg.CreatedAt),
			}
			// Отправляем клиенту
			if err := stream.Send(protoMsg); err != nil {
				if err == io.EOF {
					return nil
				}

				return status.Errorf(codes.Internal, "failed to send message: %v", err)
			}
		case <-stream.Context().Done():
			// Если клиент отключился
			return nil
		}
	}
}
