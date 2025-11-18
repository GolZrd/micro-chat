package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/config"
	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/model"
	authRepoMocks "github.com/GolZrd/micro-chat/auth/internal/repository/auth/mocks"
	authService "github.com/GolZrd/micro-chat/auth/internal/service/auth"
	jwtMocks "github.com/GolZrd/micro-chat/auth/internal/utils/jwt/mocks"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func init() {
	logger.Init(zapcore.NewNopCore())
}

func TestRefreshToken(t *testing.T) {
	refreshSecretKey := "refresh-secret-key"
	RefreshTTL := 24 * time.Hour

	testClaims := model.UserClaims{
		UID:  1,
		Name: "Test user",
		Role: "user",
	}

	tests := []struct {
		name               string
		oldRefreshToken    string
		jwtManagerMock     func(*jwtMocks.JWTManagerMock)
		authRepositoryMock func(*authRepoMocks.AuthRepositoryMock, context.Context, string)
		expectedErr        string
		expectSuccess      bool
	}{
		{
			name:            "success case",
			oldRefreshToken: "old-refresh-token",
			jwtManagerMock: func(mock *jwtMocks.JWTManagerMock) {
				mock.VerifyTokenMock.Expect("old-refresh-token", []byte(refreshSecretKey)).Return(&testClaims, nil)

				mock.GenerateTokenMock.Expect(
					model.UserAuthData{Id: testClaims.UID, Name: testClaims.Name, Role: testClaims.Role},
					refreshSecretKey,
					RefreshTTL,
				).Return("generated-refresh-token", nil)
			},
			authRepositoryMock: func(mock *authRepoMocks.AuthRepositoryMock, ctx context.Context, oldRefreshToken string) {

				mock.RevokeTokenMock.Expect(ctx, oldRefreshToken).Return(nil)

				// Set используем, т.к. в самой функции login у нас генерируется токен и expiersAt и его мы потом передаем в этот метод
				mock.CreateRefreshTokenMock.Set(func(ctx context.Context, uid int64, token string, expiresAt time.Time) error {
					require.Equal(t, testClaims.UID, uid)
					require.Equal(t, "generated-refresh-token", token)
					return nil
				})
			},
			expectSuccess: true,
		},
		{
			name:            "invalid refresh token",
			oldRefreshToken: "invalid-token",
			jwtManagerMock: func(mock *jwtMocks.JWTManagerMock) {
				mock.VerifyTokenMock.Expect("invalid-token", []byte(refreshSecretKey)).Return(nil, errors.New("invalid refresh token"))
			},
			authRepositoryMock: func(mock *authRepoMocks.AuthRepositoryMock, ctx context.Context, oldRefreshToken string) {},
			expectedErr:        "invalid refresh token",
			expectSuccess:      false,
		},
		{
			name:            "revoke tokens - non critical error",
			oldRefreshToken: "old-refresh-token",
			jwtManagerMock: func(mock *jwtMocks.JWTManagerMock) {
				mock.VerifyTokenMock.Expect("old-refresh-token", []byte(refreshSecretKey)).Return(&testClaims, nil)

				mock.GenerateTokenMock.Expect(
					model.UserAuthData{Id: testClaims.UID, Name: testClaims.Name, Role: testClaims.Role},
					refreshSecretKey,
					RefreshTTL,
				).Return("generated-refresh-token", nil)
			},
			authRepositoryMock: func(mock *authRepoMocks.AuthRepositoryMock, ctx context.Context, oldRefreshToken string) {

				mock.RevokeTokenMock.Expect(ctx, oldRefreshToken).Return(errors.New("revoke error"))

				// Set используем, т.к. в самой функции login у нас генерируется токен и expiersAt и его мы потом передаем в этот метод
				mock.CreateRefreshTokenMock.Set(func(ctx context.Context, uid int64, token string, expiresAt time.Time) error {
					require.Equal(t, testClaims.UID, uid)
					require.Equal(t, "generated-refresh-token", token)
					return nil
				})
			},
			expectSuccess: true,
		},
		{
			name:            "generate token error",
			oldRefreshToken: "test-refresh-token",
			jwtManagerMock: func(mock *jwtMocks.JWTManagerMock) {
				mock.VerifyTokenMock.Expect("test-refresh-token", []byte(refreshSecretKey)).Return(&testClaims, nil)

				mock.GenerateTokenMock.Expect(
					model.UserAuthData{Id: testClaims.UID, Name: testClaims.Name, Role: testClaims.Role},
					refreshSecretKey,
					RefreshTTL,
				).Return("", errors.New("generate token error"))
			},
			authRepositoryMock: func(mock *authRepoMocks.AuthRepositoryMock, ctx context.Context, oldRefreshToken string) {
				mock.RevokeTokenMock.Expect(ctx, oldRefreshToken).Return(nil)
			},
			expectedErr:   "generate token error",
			expectSuccess: false,
		},
		{
			name:            "save refresh token - db error",
			oldRefreshToken: "test-refresh-token",
			jwtManagerMock: func(mock *jwtMocks.JWTManagerMock) {
				mock.VerifyTokenMock.Expect("test-refresh-token", []byte(refreshSecretKey)).Return(&testClaims, nil)

				mock.GenerateTokenMock.Expect(
					model.UserAuthData{Id: testClaims.UID, Name: testClaims.Name, Role: testClaims.Role},
					refreshSecretKey,
					RefreshTTL,
				).Return("generated-refresh-token", nil)
			},
			authRepositoryMock: func(mock *authRepoMocks.AuthRepositoryMock, ctx context.Context, oldRefreshToken string) {

				mock.RevokeTokenMock.Expect(ctx, oldRefreshToken).Return(nil)

				// Set используем, т.к. в самой функции login у нас генерируется токен и expiersAt и его мы потом передаем в этот метод
				mock.CreateRefreshTokenMock.Set(func(ctx context.Context, uid int64, token string, expiresAt time.Time) error {
					require.Equal(t, testClaims.UID, uid)
					require.Equal(t, "generated-refresh-token", token)
					return errors.New("save refresh token - db error")
				})
			},
			expectedErr:   "save refresh token - db error",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)

			ctx := context.Background()

			authRepoMock := authRepoMocks.NewAuthRepositoryMock(mc)
			jwtManagerMock := jwtMocks.NewJWTManagerMock(mc)

			tt.authRepositoryMock(authRepoMock, ctx, tt.oldRefreshToken)
			tt.jwtManagerMock(jwtManagerMock)

			serviceMock := authService.NewService(authRepoMock, nil, jwtManagerMock, &config.Config{RefreshSecretKey: refreshSecretKey, RefreshTTL: RefreshTTL})

			refreshToken, err := serviceMock.RefreshToken(ctx, tt.oldRefreshToken)

			if tt.expectSuccess {
				require.NoError(t, err)
				require.NotEmpty(t, refreshToken)
				require.Equal(t, "generated-refresh-token", refreshToken)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
				require.Empty(t, refreshToken)
			}
		})
	}
}
