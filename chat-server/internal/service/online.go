package service

// OnlineUsers возвращает список онлайн пользователей в чате
func (s *service) OnlineUsers(chatID int64) []OnlineUserDTO {
	room := s.getRoom(chatID)
	if room == nil {
		return nil
	}
	return room.GetOnlineUsers()
}

// GetOnlineCount возвращает количество онлайн пользователей в чате
func (s *service) OnlineCount(chatID int64) int {
	room := s.getRoom(chatID)
	if room == nil {
		return 0
	}
	return room.GetOnlineUsersCount()
}

// IsUserOnline проверяет онлайн ли пользователь в чате
func (s *service) IsUserOnline(chatID int64, userID int64) bool {
	room := s.getRoom(chatID)
	if room == nil {
		return false
	}
	return room.IsUserOnline(userID)
}
