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

func TestRefreshToken(t *testing.T) {

	tests := []struct {
		name            string
		oldRefreshToken string
		InitMock        func(*serviceMocks.AuthServiceMock, context.Context, string)
		ExpectedCode    codes.Code
		ExpectedToken   string
		ExpectedErrMsg  string
		ExpectSuccess   bool
	}{
		{
			name:            "success case",
			oldRefreshToken: "validRefreshToken",
			InitMock: func(mock *serviceMocks.AuthServiceMock, ctx context.Context, oldRefreshToken string) {
				mock.RefreshTokenMock.Expect(ctx, oldRefreshToken).Return("newRefreshToken", nil)
			},
			ExpectedToken: "newRefreshToken",
			ExpectSuccess: true,
		},
		{
			name:            "missing old refresh token",
			oldRefreshToken: "",
			InitMock:        func(mock *serviceMocks.AuthServiceMock, ctx context.Context, oldRefreshToken string) {},
			ExpectedCode:    codes.InvalidArgument,
			ExpectedErrMsg:  "old_refresh_token is required",
			ExpectSuccess:   false,
		},
		{
			name:            "service error",
			oldRefreshToken: "ValidRefreshToken",
			InitMock: func(mock *serviceMocks.AuthServiceMock, ctx context.Context, oldRefreshToken string) {
				mock.RefreshTokenMock.Expect(ctx, oldRefreshToken).Return("", errors.New("invalid old refresh token"))
			},
			ExpectedCode:   codes.Internal,
			ExpectedErrMsg: "invalid old refresh token",
			ExpectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)

			ctx := context.Background()

			serviceMock := serviceMocks.NewAuthServiceMock(mc)
			tt.InitMock(serviceMock, ctx, tt.oldRefreshToken)

			api := auth.NewImplementation(serviceMock)

			resp, err := api.GetRefreshToken(ctx, &desc.GetRefreshTokenRequest{OldRefreshToken: tt.oldRefreshToken})

			// Проверяем, если успешный кейс
			if tt.ExpectSuccess {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tt.ExpectedToken, resp.RefreshToken)
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
