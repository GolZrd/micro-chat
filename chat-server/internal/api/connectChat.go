package api

import (
	"errors"
	"fmt"
	"io"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"github.com/GolZrd/micro-chat/chat-server/internal/utils"
	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Implementation) ConnectChat(req *desc.ConnectChatRequest, stream desc.Chat_ConnectChatServer) error {
	if req.ChatId <= 0 {
		return status.Error(codes.InvalidArgument, "chat id is required")
	}

	// Извлекаем userId из контекста
	userId, err := utils.GetUIDFromContext(stream.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, "authentication is required")
	}

	// Подписываемся на сообщения чата
	msgChan, err := s.chatService.ConnectToChat(stream.Context(), userId, req.ChatId)
	if err != nil {
		return status.Errorf(codes.NotFound, "failed to connect to chat: %v", err)
	}

	// Формируем ID подписчика
	subscriberId := fmt.Sprintf("sub_%d_%d", req.ChatId, userId)

	// При отключении от чата отписываемся
	defer s.chatService.DisconnectFromChat(req.ChatId, subscriberId)

	// Слушаем канал и отправляем сообщения клиенту
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				logger.Debug("Message channel closed", zap.Int64("chat_id", req.ChatId))
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
				if errors.Is(err, io.EOF) {
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
