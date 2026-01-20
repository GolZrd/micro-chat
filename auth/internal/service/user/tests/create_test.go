package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/model"
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

func TestCreate(t *testing.T) {
	tests := []struct {
		name               string
		input              userService.CreateUserDTO
		UserRepositoryMock func(*repoMocks.UserRepositoryMock, context.Context, userService.CreateUserDTO)
		expectedId         int64
		expectedErr        string
		expectSuccess      bool
	}{
		{
			name: "success case",
			input: userService.CreateUserDTO{
				Username:        "test",
				Email:           "test@mail.ru",
				Password:        "test",
				PasswordConfirm: "test",
				Role:            "user",
			},
			UserRepositoryMock: func(mock *repoMocks.UserRepositoryMock, ctx context.Context, input userService.CreateUserDTO) {
				mock.GetByEmailMock.Expect(ctx, input.Email).Return(nil, userRepository.ErrUserNotFound)
				mock.CreateMock.Expect(ctx, userRepository.CreateUserDTO{
					Username: input.Username,
					Email:    input.Email,
					Password: input.Password,
					Role:     input.Role,
				}).Return(int64(52), nil)
			},
			expectedId:    52,
			expectSuccess: true,
		},
		{
			name: "passwords do not match",
			input: userService.CreateUserDTO{
				Username:        "test",
				Email:           "test@mail.ru",
				Password:        "test",
				PasswordConfirm: "Test2",
				Role:            "user",
			},
			UserRepositoryMock: func(mock *repoMocks.UserRepositoryMock, ctx context.Context, input userService.CreateUserDTO) {},
			expectedErr:        "passwords do not match",
			expectSuccess:      false,
		},
		{
			name: "user already exists",
			input: userService.CreateUserDTO{
				Username:        "test",
				Email:           "test@mail.ru",
				Password:        "test",
				PasswordConfirm: "test",
				Role:            "user",
			},
			UserRepositoryMock: func(mock *repoMocks.UserRepositoryMock, ctx context.Context, input userService.CreateUserDTO) {
				mock.GetByEmailMock.Expect(ctx, input.Email).Return(&model.UserAuthData{Id: 52}, nil)

			},
			expectedErr:   "user already exists",
			expectSuccess: false,
		},
		{
			name: "get user by email - db error",
			input: userService.CreateUserDTO{
				Username:        "test",
				Email:           "test@mail.ru",
				Password:        "test",
				PasswordConfirm: "test",
				Role:            "user",
			},
			UserRepositoryMock: func(mock *repoMocks.UserRepositoryMock, ctx context.Context, input userService.CreateUserDTO) {
				mock.GetByEmailMock.Expect(ctx, input.Email).Return(nil, errors.New("get user by email - db error"))
			},
			expectedErr:   "get user by email - db error",
			expectSuccess: false,
		},
		{
			name: "create user - db error",
			input: userService.CreateUserDTO{
				Username:        "test",
				Email:           "test@mail.ru",
				Password:        "test",
				PasswordConfirm: "test",
				Role:            "user",
			},
			UserRepositoryMock: func(mock *repoMocks.UserRepositoryMock, ctx context.Context, input userService.CreateUserDTO) {
				mock.GetByEmailMock.Expect(ctx, input.Email).Return(nil, userRepository.ErrUserNotFound)
				mock.CreateMock.Expect(ctx, userRepository.CreateUserDTO{
					Username: input.Username,
					Email:    input.Email,
					Password: input.Password,
					Role:     input.Role,
				}).Return(0, errors.New("create user - db error"))
			},
			expectedErr:   "create user - db error",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)

			ctx := context.Background()

			repoMock := repoMocks.NewUserRepositoryMock(mc)
			tt.UserRepositoryMock(repoMock, ctx, tt.input)

			serviceMock := userService.NewService(repoMock)

			id, err := serviceMock.Create(ctx, tt.input)

			if tt.expectSuccess {
				require.NoError(t, err)
				require.Equal(t, tt.expectedId, id)
			} else {
				require.Error(t, err)
				require.Equal(t, int64(0), id)
				require.Contains(t, err.Error(), tt.expectedErr)
			}
		})
	}
}
