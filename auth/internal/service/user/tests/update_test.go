package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	userRepository "github.com/GolZrd/micro-chat/auth/internal/repository/user"
	repoMocks "github.com/GolZrd/micro-chat/auth/internal/repository/user/mocks"
	userService "github.com/GolZrd/micro-chat/auth/internal/service/user"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

// Инициализируем logger перед запуском тестов c помощью NopCore
func init() {
	logger.Init(zapcore.NewNopCore())
}

func TestUpdate(t *testing.T) {
	usernameToUpdate := "test"
	emailToUpdate := "test@mail.ru"

	tests := []struct {
		name               string
		userId             int64
		input              userService.UpdateUserDTO
		UserRepositoryMock func(*repoMocks.UserRepositoryMock, context.Context, int64, userService.UpdateUserDTO)
		expectedErr        string
		expectSuccess      bool
	}{
		{
			name:   "success case - update only name",
			userId: int64(1),
			input: userService.UpdateUserDTO{
				Username: &usernameToUpdate,
				Email:    nil,
			},
			UserRepositoryMock: func(mock *repoMocks.UserRepositoryMock, ctx context.Context, userId int64, input userService.UpdateUserDTO) {
				mock.UpdateMock.Expect(ctx, userId, userRepository.UpdateUserDTO{
					Username: input.Username,
					Email:    input.Email,
				}).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:   "success case - update only email",
			userId: int64(1),
			input: userService.UpdateUserDTO{
				Username: nil,
				Email:    &emailToUpdate,
			},
			UserRepositoryMock: func(mock *repoMocks.UserRepositoryMock, ctx context.Context, userId int64, input userService.UpdateUserDTO) {
				mock.UpdateMock.Expect(ctx, userId, userRepository.UpdateUserDTO{
					Username: input.Username,
					Email:    input.Email,
				}).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:   "success case - update both name and email",
			userId: int64(1),
			input: userService.UpdateUserDTO{
				Username: &usernameToUpdate,
				Email:    &emailToUpdate,
			},
			UserRepositoryMock: func(mock *repoMocks.UserRepositoryMock, ctx context.Context, userId int64, input userService.UpdateUserDTO) {
				mock.UpdateMock.Expect(ctx, userId, userRepository.UpdateUserDTO{
					Username: input.Username,
					Email:    input.Email,
				}).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:   "failed to update user - db error",
			userId: int64(1),
			input: userService.UpdateUserDTO{
				Username: &usernameToUpdate,
				Email:    &emailToUpdate,
			},
			UserRepositoryMock: func(mock *repoMocks.UserRepositoryMock, ctx context.Context, userId int64, input userService.UpdateUserDTO) {
				mock.UpdateMock.Expect(ctx, userId, userRepository.UpdateUserDTO{
					Username: input.Username,
					Email:    input.Email,
				}).Return(errors.New("failed to update user - db error"))
			},
			expectedErr:   "failed to update user - db error",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)

			ctx := context.Background()

			repoMock := repoMocks.NewUserRepositoryMock(mc)
			tt.UserRepositoryMock(repoMock, ctx, tt.userId, tt.input)

			serviceMock := userService.NewService(repoMock)

			err := serviceMock.Update(ctx, tt.userId, tt.input)

			if tt.expectSuccess {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
			}
		})
	}
}
