package service

import (
	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
)

func (s *service) DisconnectFromChat(chatId int64, userId int64) {
	room := s.getRoom(chatId)
	if room == nil {
		return
	}

	// Удаляем подписчика из комнаты
	channel, isEmpty := room.RemoveSubscriber(userId)

	if channel == nil {
		return // Подписчик не был в комнате
	}

	// Закрываем канал подписчика
	close(channel)

	logger.Info("subscriber disconnected", zap.Int64("chat_id", chatId), zap.Int64("user_id", userId), zap.Int("remaining_online", room.GetOnlineUsersCount()))

	// Если комната пуста, то удаляем ее
	if isEmpty {
		s.deleteRoomIfEmpty(chatId)
		logger.Debug("no subscribers left in chat", zap.Int64("chat_id", chatId))
	} else {
		// Если не пустая, то уведомляем об актуальном онлайне в чате
		go room.BroadcastOnlineUsers()
	}
}
