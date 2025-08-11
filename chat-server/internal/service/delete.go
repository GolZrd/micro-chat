package service

import (
	"context"
	"log"
)

func (s *service) Delete(ctx context.Context, id int64) error {
	err := s.ChatRepository.Delete(ctx, id)
	if err != nil {
		return err
	}

	log.Printf("Deleted chat with id: %d", id)

	return nil
}
