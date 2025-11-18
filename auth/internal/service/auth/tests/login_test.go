package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/config"
	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/model"
	authRepo "github.com/GolZrd/micro-chat/auth/internal/repository/auth"
	authRepoMocks "github.com/GolZrd/micro-chat/auth/internal/repository/auth/mocks"
	userRepoMocks "github.com/GolZrd/micro-chat/auth/internal/repository/user/mocks"
	authService "github.com/GolZrd/micro-chat/auth/internal/service/auth"
	jwtMocks "github.com/GolZrd/micro-chat/auth/internal/utils/jwt/mocks"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func init() {
	logger.Init(zapcore.NewNopCore())
}

func TestLogin(t *testing.T) {
	secretKey := "test-secret-key"
	RefreshTTL := 24 * time.Hour

	userData := &model.UserAuthData{
		Id:       1,
		Name:     "Test user",
		Email:    "test@mail.ru",
		Password: "test-password",
		Role:     "user",
	}

	tests := []struct {
		name               string
		email              string
		password           string
		userRepositoryMock func(*userRepoMocks.UserRepositoryMock, context.Context, string)
		authRepositoryMock func(*authRepoMocks.AuthRepositoryMock, context.Context, int64)
		jwtManagerMock     func(*jwtMocks.JWTManagerMock)
		expectedUserId     int64
		expectedErr        string
		expectSuccess      bool
	}{
		{
			name:     "success case",
			email:    "Test@mail.ru",
			password: "test-password",
			userRepositoryMock: func(mock *userRepoMocks.UserRepositoryMock, ctx context.Context, email string) {
				mock.GetByEmailMock.Expect(ctx, userData.Email).Return(userData, nil)
			},
			authRepositoryMock: func(mock *authRepoMocks.AuthRepositoryMock, ctx context.Context, userId int64) {
				mock.RevokeAllByUserIdMock.Expect(ctx, userId).Return(nil)

				// Set используем, т.к. в самой функции login у нас генерируется токен и expiersAt и его мы потом передаем в этот метод
				mock.CreateRefreshTokenMock.Set(func(ctx context.Context, uid int64, token string, expiresAt time.Time) error {
					require.Equal(t, userId, uid)
					// Проверяем что токен создался и не пустой
					require.NotEmpty(t, token)
					return nil
				})
			},
			jwtManagerMock: func(mock *jwtMocks.JWTManagerMock) {
				mock.GenerateTokenMock.Expect(
					model.UserAuthData{Id: userData.Id, Name: userData.Name, Role: userData.Role},
					secretKey,
					RefreshTTL,
				).Return("generated-refresh-token", nil)
			},
			expectedUserId: int64(1),
			expectSuccess:  true,
		},
		{
			name:     "user not found",
			email:    "Test@mail.ru",
			password: "test-password",
			userRepositoryMock: func(mock *userRepoMocks.UserRepositoryMock, ctx context.Context, email string) {
				mock.GetByEmailMock.Expect(ctx, userData.Email).Return(nil, authRepo.ErrUserNotFound)
			},
			authRepositoryMock: func(mock *authRepoMocks.AuthRepositoryMock, ctx context.Context, userId int64) {
			},
			jwtManagerMock: func(jm *jwtMocks.JWTManagerMock) {},
			expectedErr:    "invalid credentials",
			expectSuccess:  false,
		},
		{
			name:     "get user by email - db error",
			email:    "Test@mail.ru",
			password: "test-password",
			userRepositoryMock: func(mock *userRepoMocks.UserRepositoryMock, ctx context.Context, email string) {
				mock.GetByEmailMock.Expect(ctx, userData.Email).Return(nil, errors.New("get user by email - db error"))
			},
			authRepositoryMock: func(mock *authRepoMocks.AuthRepositoryMock, ctx context.Context, userId int64) {},
			jwtManagerMock:     func(jm *jwtMocks.JWTManagerMock) {},
			expectedErr:        "get user by email - db error",
			expectSuccess:      false,
		},
		{
			name:     "invalid password",
			email:    "Test@mail.ru",
			password: "wrong-password",
			userRepositoryMock: func(mock *userRepoMocks.UserRepositoryMock, ctx context.Context, email string) {
				mock.GetByEmailMock.Expect(ctx, userData.Email).Return(userData, nil)
			},
			authRepositoryMock: func(mock *authRepoMocks.AuthRepositoryMock, ctx context.Context, userId int64) {},
			jwtManagerMock:     func(jm *jwtMocks.JWTManagerMock) {},
			expectedErr:        "invalid credentials",
			expectSuccess:      false,
		},
		{
			name:     "revoke all tokens - non critical error",
			email:    "Test@mail.ru",
			password: "test-password",
			userRepositoryMock: func(mock *userRepoMocks.UserRepositoryMock, ctx context.Context, email string) {
				mock.GetByEmailMock.Expect(ctx, userData.Email).Return(userData, nil)
			},
			authRepositoryMock: func(mock *authRepoMocks.AuthRepositoryMock, ctx context.Context, userId int64) {
				// Возвращает некритическую ошибку, поэтому продолжаем выполнение
				mock.RevokeAllByUserIdMock.Expect(ctx, userId).Return(errors.New("revoke error"))

				mock.CreateRefreshTokenMock.Set(func(ctx context.Context, uid int64, token string, expiresAt time.Time) error {
					require.Equal(t, userId, uid)
					require.NotEmpty(t, token)
					return nil
				})
			},
			jwtManagerMock: func(mock *jwtMocks.JWTManagerMock) {
				mock.GenerateTokenMock.Expect(
					model.UserAuthData{Id: userData.Id, Name: userData.Name, Role: userData.Role},
					secretKey,
					RefreshTTL,
				).Return("generated-refresh-token", nil)
			},
			expectedUserId: int64(1),
			expectSuccess:  true,
		},
		{
			name:     "save refresh token - db error",
			email:    "Test@mail.ru",
			password: "test-password",
			userRepositoryMock: func(mock *userRepoMocks.UserRepositoryMock, ctx context.Context, email string) {
				mock.GetByEmailMock.Expect(ctx, userData.Email).Return(userData, nil)
			},
			authRepositoryMock: func(mock *authRepoMocks.AuthRepositoryMock, ctx context.Context, userId int64) {
				mock.RevokeAllByUserIdMock.Expect(ctx, userId).Return(nil)

				mock.CreateRefreshTokenMock.Set(func(ctx context.Context, uid int64, token string, expiresAt time.Time) error {
					require.Equal(t, userId, uid)
					// Проверяем что токен создался и не пустой
					require.NotEmpty(t, token)
					return errors.New("save refresh token - db error")
				})
			},
			jwtManagerMock: func(mock *jwtMocks.JWTManagerMock) {
				mock.GenerateTokenMock.Expect(
					model.UserAuthData{Id: userData.Id, Name: userData.Name, Role: userData.Role},
					secretKey,
					RefreshTTL,
				).Return("generated-refresh-token", nil)
			},
			expectedUserId: int64(1),
			expectedErr:    "save refresh token - db error",
			expectSuccess:  false,
		},
		{
			name:     "generate token error",
			email:    "Test@mail.ru",
			password: "test-password",
			userRepositoryMock: func(mock *userRepoMocks.UserRepositoryMock, ctx context.Context, email string) {
				mock.GetByEmailMock.Expect(ctx, userData.Email).Return(userData, nil)
			},
			authRepositoryMock: func(mock *authRepoMocks.AuthRepositoryMock, ctx context.Context, userId int64) {
				mock.RevokeAllByUserIdMock.Expect(ctx, userId).Return(nil)

				// Второй метод не вызывается, потому что ошибка генерации токена
			},
			jwtManagerMock: func(mock *jwtMocks.JWTManagerMock) {
				mock.GenerateTokenMock.Expect(
					model.UserAuthData{Id: userData.Id, Name: userData.Name, Role: userData.Role},
					secretKey,
					RefreshTTL,
				).Return("", errors.New("jwt generation failed"))
			},
			expectedUserId: int64(1),
			expectedErr:    "jwt generation failed",
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)

			ctx := context.Background()

			userRepoMock := userRepoMocks.NewUserRepositoryMock(mc)
			authRepoMock := authRepoMocks.NewAuthRepositoryMock(mc)
			jwtManagerMock := jwtMocks.NewJWTManagerMock(mc)

			tt.userRepositoryMock(userRepoMock, ctx, tt.email)
			tt.authRepositoryMock(authRepoMock, ctx, tt.expectedUserId)
			tt.jwtManagerMock(jwtManagerMock)

			serviceMock := authService.NewService(authRepoMock, userRepoMock, jwtManagerMock, &config.Config{RefreshSecretKey: secretKey, RefreshTTL: RefreshTTL})

			refreshToken, userId, err := serviceMock.Login(ctx, tt.email, tt.password)

			if tt.expectSuccess {
				require.NoError(t, err)
				require.NotNil(t, refreshToken)
				require.Equal(t, tt.expectedUserId, userId)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
				require.Empty(t, refreshToken)
			}
		})
	}
}
