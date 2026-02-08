package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/GolZrd/micro-chat/auth/internal/api/user"
	userService "github.com/GolZrd/micro-chat/auth/internal/service/user"
	serviceMocks "github.com/GolZrd/micro-chat/auth/internal/service/user/mocks"
	desc "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		req            *desc.CreateRequest
		InitMock       func(*serviceMocks.UserServiceMock, context.Context, userService.CreateUserDTO)
		expectedId     int64
		expectedCode   codes.Code
		expectedErrMsg string
		expectSuccess  bool
	}{
		{
			name: "success case",
			req: &desc.CreateRequest{
				Info: &desc.UserInfo{
					Username:        "test",
					Email:           "test@mail.ru",
					Password:        "test",
					PasswordConfirm: "test",
					Role:            desc.Role_user,
				},
			},
			InitMock: func(mock *serviceMocks.UserServiceMock, ctx context.Context, dto userService.CreateUserDTO) {
				mock.CreateMock.Expect(ctx, dto).Return(int64(64), nil)
			},
			expectedId:    64,
			expectSuccess: true,
		},
		{
			name: "validation error - info is nil",
			req: &desc.CreateRequest{
				Info: nil,
			},
			InitMock:       func(mock *serviceMocks.UserServiceMock, ctx context.Context, dto userService.CreateUserDTO) {},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "info is required",
			expectSuccess:  false,
		},
		{
			name: "validation error - name is empty",
			req: &desc.CreateRequest{
				Info: &desc.UserInfo{
					Username:        "",
					Email:           "test@mail.ru",
					Password:        "test",
					PasswordConfirm: "test",
					Role:            desc.Role_user,
				},
			},
			InitMock:       func(mock *serviceMocks.UserServiceMock, ctx context.Context, dto userService.CreateUserDTO) {},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "name is required",
			expectSuccess:  false,
		},
		{
			name: "validation error - email is empty",
			req: &desc.CreateRequest{
				Info: &desc.UserInfo{
					Username:        "test",
					Email:           "",
					Password:        "test",
					PasswordConfirm: "test",
					Role:            desc.Role_user,
				},
			},
			InitMock:       func(mock *serviceMocks.UserServiceMock, ctx context.Context, dto userService.CreateUserDTO) {},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "email is required",
			expectSuccess:  false,
		},
		{
			name: "validation error - password is empty",
			req: &desc.CreateRequest{
				Info: &desc.UserInfo{
					Username:        "test",
					Email:           "test@mail.ru",
					Password:        "",
					PasswordConfirm: "test",
					Role:            desc.Role_user,
				},
			},
			InitMock:       func(mock *serviceMocks.UserServiceMock, ctx context.Context, dto userService.CreateUserDTO) {},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "password is required",
			expectSuccess:  false,
		},
		{
			name: "validation error - password confirm is empty",
			req: &desc.CreateRequest{
				Info: &desc.UserInfo{
					Username:        "test",
					Email:           "test@mail.ru",
					Password:        "test",
					PasswordConfirm: "",
					Role:            desc.Role_user,
				},
			},
			InitMock:       func(mock *serviceMocks.UserServiceMock, ctx context.Context, dto userService.CreateUserDTO) {},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "password_confirm is required",
			expectSuccess:  false,
		},
		{
			name: "service error",
			req: &desc.CreateRequest{
				Info: &desc.UserInfo{
					Username:        "test",
					Email:           "test@mail.ru",
					Password:        "test",
					PasswordConfirm: "test",
					Role:            desc.Role_user,
				},
			},
			InitMock: func(mock *serviceMocks.UserServiceMock, ctx context.Context, dto userService.CreateUserDTO) {
				mock.CreateMock.Expect(ctx, dto).Return(0, errors.New("service error"))
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

			serviceMock := serviceMocks.NewUserServiceMock(mc)

			if tt.req.Info != nil {
				tt.InitMock(serviceMock, ctx, userService.CreateUserDTO{
					Username:        tt.req.Info.Username,
					Email:           tt.req.Info.Email,
					Password:        tt.req.Info.Password,
					PasswordConfirm: tt.req.Info.PasswordConfirm,
					Role:            tt.req.Info.Role.String(),
				})

				api := user.NewImplementation(serviceMock, nil, nil) // jwtManager, cfg

				resp, err := api.Create(ctx, tt.req)

				if tt.expectSuccess {
					require.NoError(t, err)
					require.NotNil(t, resp)
					require.Equal(t, tt.expectedId, resp.Id)
				} else {
					require.Error(t, err)
					require.Nil(t, resp)
					require.Equal(t, tt.expectedCode, status.Code(err))
					require.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			}
		})
	}

}
