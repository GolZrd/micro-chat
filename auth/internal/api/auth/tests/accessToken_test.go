package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/GolZrd/micro-chat/auth/internal/api/auth"
	serviceMocks "github.com/GolZrd/micro-chat/auth/internal/service/auth/mocks"
	desc "github.com/GolZrd/micro-chat/auth/pkg/auth_v1"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccessToken(t *testing.T) {

	tests := []struct {
		name           string
		refreshToken   string
		InitMock       func(*serviceMocks.AuthServiceMock, context.Context, string)
		ExpectedCode   codes.Code
		ExpectedToken  string
		ExpectedErrMsg string
		ExpectSuccess  bool
	}{
		{
			name:         "success case",
			refreshToken: "validRefreshToken",
			InitMock: func(mock *serviceMocks.AuthServiceMock, ctx context.Context, refreshToken string) {
				mock.AccessTokenMock.Expect(ctx, refreshToken).Return("newAccessToken", nil)
			},
			ExpectedToken: "newAccessToken",
			ExpectSuccess: true,
		},
		{
			name:           "missing refresh token",
			refreshToken:   "",
			InitMock:       func(mock *serviceMocks.AuthServiceMock, ctx context.Context, refreshToken string) {},
			ExpectedCode:   codes.InvalidArgument,
			ExpectedErrMsg: "refresh_token is required",
			ExpectSuccess:  false,
		},
		{
			name:         "service error",
			refreshToken: "ValidRefreshToken",
			InitMock: func(mock *serviceMocks.AuthServiceMock, ctx context.Context, refreshToken string) {
				mock.AccessTokenMock.Expect(ctx, refreshToken).Return("", errors.New("invalid refresh token"))
			},
			ExpectedCode:   codes.Internal,
			ExpectedErrMsg: "invalid refresh token",
			ExpectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)

			ctx := context.Background()

			serviceMock := serviceMocks.NewAuthServiceMock(mc)
			tt.InitMock(serviceMock, ctx, tt.refreshToken)

			api := auth.NewImplementation(serviceMock)

			resp, err := api.GetAccessToken(ctx, &desc.GetAccessTokenRequest{RefreshToken: tt.refreshToken})

			// Проверяем, если успешный кейс
			if tt.ExpectSuccess {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tt.ExpectedToken, resp.AccessToken)
			} else {
				require.Error(t, err)
				require.Nil(t, resp)

				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.ExpectedCode, st.Code())
				require.Contains(t, st.Message(), tt.ExpectedErrMsg)
			}
		})
	}
}
