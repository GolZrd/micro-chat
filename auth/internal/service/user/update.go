package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	userRepository "github.com/GolZrd/micro-chat/auth/internal/repository/user"
	"go.uber.org/zap"
)

func (s *service) Update(ctx context.Context, id int64, info UpdateUserDTO) error {
	params := userRepository.UpdateUserDTO{}
	// Срез для проверки, что хотя бы одно поле заполнено
	fieldsToUpdate := []string{}

	// Проверяем что хотя бы одно поле заполнено
	if info.Name != nil {
		// Убираем лишние пробелы
		name := strings.TrimSpace(*info.Name)

		// TODO: добавить валидацию имени

		// Добавляем в DTO
		params.Name = &name
		fieldsToUpdate = append(fieldsToUpdate, "name")
	}

	if info.Email != nil {
		// Убираем лишние пробелы и переводим в нижний регистр
		email := strings.TrimSpace(strings.ToLower(*info.Email))

		// TODO: добавить валидацию email

		// Добавляем в DTO
		params.Email = &email
		fieldsToUpdate = append(fieldsToUpdate, "email")
	}

	err := s.userRepository.Update(ctx, id, params)
	if err != nil {
		logger.Error("Failed to update user in DB", zap.Int64("user_id", id), zap.Strings("fields", fieldsToUpdate), zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info("user updated successfully",
		zap.Int64("user_id", id),
		zap.Strings("fields", fieldsToUpdate),
	)

	return nil
}
