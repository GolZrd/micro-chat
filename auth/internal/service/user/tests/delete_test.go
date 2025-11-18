package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
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

func TestDelete(t *testing.T) {
	tests := []struct {
		name               string
		userId             int64
		UserRepositoryMock func(*repoMocks.UserRepositoryMock, context.Context, int64)
		expectedErr        string
		expectSuccess      bool
	}{
		{
			name:   "success case",
			userId: 1,
			UserRepositoryMock: func(mock *repoMocks.UserRepositoryMock, ctx context.Context, userId int64) {
				mock.DeleteMock.Expect(ctx, userId).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:   "delete user - db error",
			userId: 23,
			UserRepositoryMock: func(mock *repoMocks.UserRepositoryMock, ctx context.Context, userId int64) {
				mock.DeleteMock.Expect(ctx, userId).Return(errors.New("delete user - db error"))
			},
			expectedErr:   "delete user - db error",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)

			ctx := context.Background()

			repoMock := repoMocks.NewUserRepositoryMock(mc)
			tt.UserRepositoryMock(repoMock, ctx, tt.userId)

			serviceMock := userService.NewService(repoMock)

			err := serviceMock.Delete(ctx, tt.userId)

			if tt.expectSuccess {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
			}
		})
	}
}
