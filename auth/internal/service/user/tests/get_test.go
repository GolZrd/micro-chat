package tests

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/model"
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

func TestGet(t *testing.T) {
	userData := &model.User{
		Id:        64,
		Info:      model.UserInfo{},
		CreatedAt: time.Now(),
		UpdatedAt: sql.NullTime{},
	}

	tests := []struct {
		name               string
		userId             int64
		UserRepositoryMock func(*repoMocks.UserRepositoryMock, context.Context, int64)
		expectedUser       *model.User
		expectedErr        string
		expectSuccess      bool
	}{
		{
			name:   "success case",
			userId: int64(64),
			UserRepositoryMock: func(mock *repoMocks.UserRepositoryMock, ctx context.Context, userId int64) {
				mock.GetMock.Expect(ctx, userId).Return(userData, nil)
			},
			expectedUser:  userData,
			expectSuccess: true,
		},
		{
			name:   "get user by id - db error",
			userId: int64(64),
			UserRepositoryMock: func(mock *repoMocks.UserRepositoryMock, ctx context.Context, userId int64) {
				mock.GetMock.Expect(ctx, userId).Return(nil, errors.New("get user by id - db error"))
			},
			expectedErr:   "get user by id - db error",
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

			user, err := serviceMock.Get(ctx, tt.userId)

			if tt.expectSuccess {
				require.NoError(t, err)
				require.Equal(t, tt.expectedUser, user)
			} else {
				require.Error(t, err)
				require.Nil(t, user)
				require.Contains(t, err.Error(), tt.expectedErr)
			}
		})
	}
}
