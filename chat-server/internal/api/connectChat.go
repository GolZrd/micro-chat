package api

import (
	"errors"
	"io"

	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"github.com/GolZrd/micro-chat/chat-server/internal/service"
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

	// Извлекаем username из контекста
	username, err := utils.GetUsernameFromContext(stream.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, "authentication is required")
	}

	// Подписываемся на сообщения чата
	msgChan, err := s.chatService.ConnectToChat(stream.Context(), userId, username, req.ChatId)
	if err != nil {
		return status.Errorf(codes.NotFound, "failed to connect to chat: %v", err)
	}

	// При отключении от чата отписываемся
	defer s.chatService.DisconnectFromChat(req.ChatId, userId)

	// Слушаем канал и отправляем сообщения клиенту
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				logger.Debug("Message channel closed", zap.Int64("chat_id", req.ChatId))
				return nil
			}

			protoMsg := s.convertToProto(msg)

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

// convertToProto конвертирует MessageDTO в proto Message
func (s *Implementation) convertToProto(msg service.MessageDTO) *desc.Message {
	// В зависимости от типа сообщения собираем стрктуру proto
	switch msg.Type {
	case service.MessageTypeOnlineUsers:
		// Получаем онлайн пользователей
		onlineUsers := make([]*desc.OnlineUsers, 0, len(msg.OnlineUsers))
		for _, user := range msg.OnlineUsers {
			onlineUsers = append(onlineUsers, &desc.OnlineUsers{
				UserId:   user.UserId,
				Username: user.Username,
			})
		}

		return &desc.Message{
			Type:        desc.MessageType_MESSAGE_TYPE_ONLINE_USERS,
			OnlineUsers: onlineUsers,
			CreatedAt:   timestamppb.New(msg.CreatedAt),
		}

	default:
		// Текстовое сообщение
		return &desc.Message{
			Type:      desc.MessageType_MESSAGE_TYPE_TEXT,
			From:      msg.From,
			Text:      msg.Text,
			CreatedAt: timestamppb.New(msg.CreatedAt),
		}
	}

}
