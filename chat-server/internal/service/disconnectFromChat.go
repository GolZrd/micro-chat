package service

import (
	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

func (s *service) DisconnectFromChat(chatID int64, subscriberId string) {
	s.subMutex.Lock()
	defer s.subMutex.Unlock()

	if chatSubs, exists := s.subscribers[chatID]; exists {
		if ch, exists := chatSubs[subscriberId]; exists {
			close(ch)
			delete(chatSubs, subscriberId)

			logger.Info("subscriber disconnected", zap.Int64("chat_id", chatID), zap.String("subscriber_id", subscriberId))
		}

		if len(chatSubs) == 0 {
			delete(s.subscribers, chatID)
			logger.Debug("no subscribers left in chat", zap.Int64("chat_id", chatID))
		}
	}
}
