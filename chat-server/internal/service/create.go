package service

import (
	"context"
	"errors"
)

func (s *service) Create(ctx context.Context, usernames []string) (int64, error) {
	// Проверяем что Usernames не пусто
	if len(usernames) == 0 {
		return 0, errors.New("usernames cannot be empty")
	}

	// Пока будем просто будем создавать чат с переданными usernames, дальше нужно будет изменить
	// TODO: первым делом нужно будет проверить что пользователи существуют, с помощью запроса к сервису AUTH, где у нас регистрируются и создаются пользователи

	id, err := s.ChatRepository.Create(ctx, usernames)
	if err != nil {
		return 0, err
	}

	return id, nil
}
