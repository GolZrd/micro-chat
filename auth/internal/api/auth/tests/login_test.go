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

func TestLogin(t *testing.T) {
	serviceErr := errors.New("service error")

	tests := []struct {
		name                 string
		email                string
		password             string
		initMock             func(*serviceMocks.AuthServiceMock, context.Context, string, string)
		expectedRefreshToken string
		expectedUserId       int64
		expectedCode         codes.Code
		expectedErrMsg       string
		expectSuccess        bool
	}{
		{
			name:     "success case",
			email:    "success@mail.ru",
			password: "successPassword",
			initMock: func(mock *serviceMocks.AuthServiceMock, ctx context.Context, email string, password string) {
				mock.LoginMock.Expect(ctx, email, password).Return("validRefreshToken", int64(1), nil)
			},
			expectedRefreshToken: "validRefreshToken",
			expectedUserId:       int64(1),
			expectSuccess:        true,
		},
		{
			name:           "missing email",
			email:          "",
			password:       "password123",
			initMock:       func(mock *serviceMocks.AuthServiceMock, ctx context.Context, email string, password string) {},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "email is required",
			expectSuccess:  false,
		},
		{
			name:           "missing password",
			email:          "user@mail.ru",
			password:       "",
			initMock:       func(mock *serviceMocks.AuthServiceMock, ctx context.Context, email string, password string) {},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "password is required",
			expectSuccess:  false,
		},
		{
			name:     "service error",
			email:    "user@mail.ru",
			password: "password123",
			initMock: func(mock *serviceMocks.AuthServiceMock, ctx context.Context, email string, password string) {
				mock.LoginMock.Expect(ctx, email, password).Return("", int64(0), serviceErr)
			},
			expectedCode:   codes.Internal,
			expectedErrMsg: "service error",
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)

			ctx := context.Background()

			serviceMock := serviceMocks.NewAuthServiceMock(mc)
			tt.initMock(serviceMock, ctx, tt.email, tt.password)

			api := auth.NewImplementation(serviceMock)

			resp, err := api.Login(ctx, &desc.LoginRequest{Email: tt.email, Password: tt.password})

			if tt.expectSuccess {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tt.expectedRefreshToken, resp.RefreshToken)
				require.Equal(t, tt.expectedUserId, resp.UserId)
			} else {
				require.Error(t, err)
				require.Nil(t, resp)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.expectedCode, st.Code())
				require.Contains(t, st.Message(), tt.expectedErrMsg)
			}
		})
	}
}
