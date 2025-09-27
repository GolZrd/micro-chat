package service

func (s *service) DisconnectFromChat(chatID int64, subscriberId string) {
	s.subMutex.Lock()
	defer s.subMutex.Unlock()

	if chatSubs, exists := s.subscribers[chatID]; exists {
		if ch, exists := chatSubs[subscriberId]; exists {
			close(ch)
			delete(chatSubs, subscriberId)
		}

		if len(chatSubs) == 0 {
			delete(s.subscribers, chatID)
		}
	}
}
