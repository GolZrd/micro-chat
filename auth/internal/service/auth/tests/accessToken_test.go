package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/config"
	"github.com/GolZrd/micro-chat/auth/internal/logger"
	"github.com/GolZrd/micro-chat/auth/internal/model"
	authService "github.com/GolZrd/micro-chat/auth/internal/service/auth"
	jwtMocks "github.com/GolZrd/micro-chat/auth/internal/utils/jwt/mocks"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func init() {
	logger.Init(zapcore.NewNopCore())
}

func TestAccessToken(t *testing.T) {
	accessSecretKey := "access-secret-key"
	refreshSecretKey := "refresh-secret-key"
	accessTTL := 15 * time.Minute

	testClaims := model.UserClaims{
		UID:      1,
		Username: "Test user",
		Role:     "user",
	}

	tests := []struct {
		name           string
		refreshToken   string
		jwtManagerMock func(*jwtMocks.JWTManagerMock)
		expectedErr    string
		expectSuccess  bool
	}{
		{
			name:         "success case",
			refreshToken: "test-refresh-token",
			jwtManagerMock: func(mock *jwtMocks.JWTManagerMock) {
				mock.VerifyTokenMock.Expect("test-refresh-token", []byte(refreshSecretKey)).Return(&testClaims, nil)

				mock.GenerateTokenMock.Expect(
					model.UserAuthData{Id: testClaims.UID, Username: testClaims.Username, Role: testClaims.Role},
					accessSecretKey,
					accessTTL,
				).Return("generated-access-token", nil)
			},
			expectSuccess: true,
		},
		{
			name:         "invalid refresh token",
			refreshToken: "invalid-token",
			jwtManagerMock: func(mock *jwtMocks.JWTManagerMock) {
				mock.VerifyTokenMock.Expect("invalid-token", []byte(refreshSecretKey)).Return(nil, errors.New("invalid refresh token"))
			},
			expectedErr:   "invalid refresh token",
			expectSuccess: false,
		},
		{
			name:         "generate token error",
			refreshToken: "test-refresh-token",
			jwtManagerMock: func(mock *jwtMocks.JWTManagerMock) {
				mock.VerifyTokenMock.Expect("test-refresh-token", []byte(refreshSecretKey)).Return(&testClaims, nil)

				mock.GenerateTokenMock.Expect(
					model.UserAuthData{Id: testClaims.UID, Username: testClaims.Username, Role: testClaims.Role},
					accessSecretKey,
					accessTTL,
				).Return("", errors.New("generate token error"))
			},
			expectedErr:   "failed to generate access token",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)

			ctx := context.Background()

			jwtManagerMock := jwtMocks.NewJWTManagerMock(mc)

			tt.jwtManagerMock(jwtManagerMock)

			serviceMock := authService.NewService(nil, nil, jwtManagerMock, &config.Config{RefreshSecretKey: refreshSecretKey, AccessSecretKey: accessSecretKey, AccessTTL: accessTTL})

			accessToken, err := serviceMock.AccessToken(ctx, tt.refreshToken)

			if tt.expectSuccess {
				require.NoError(t, err)
				require.NotNil(t, accessToken)
				require.Equal(t, "generated-access-token", accessToken)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
				require.Empty(t, accessToken)
			}
		})
	}
}
