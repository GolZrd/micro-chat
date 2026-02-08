package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/GolZrd/micro-chat/auth/internal/api/user"
	"github.com/GolZrd/micro-chat/auth/internal/config"
	userService "github.com/GolZrd/micro-chat/auth/internal/service/user"
	serviceMocks "github.com/GolZrd/micro-chat/auth/internal/service/user/mocks"
	desc "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestUpdate(t *testing.T) {
	tests := []struct {
		name           string
		req            *desc.UpdateRequest
		InitMock       func(*serviceMocks.UserServiceMock, context.Context, int64, userService.UpdateUserDTO)
		expectedCode   codes.Code
		expectedErrMsg string
		expectSuccess  bool
	}{
		{
			name: "success case - update only name",
			req: &desc.UpdateRequest{
				Id: 64,
				Info: &desc.UpdateUserInfo{
					Username: wrapperspb.String("New name"),
					Email:    nil,
				},
			},
			InitMock: func(mock *serviceMocks.UserServiceMock, ctx context.Context, userId int64, dto userService.UpdateUserDTO) {
				mock.UpdateMock.Expect(ctx, userId, dto).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name: "success case - update only email",
			req: &desc.UpdateRequest{
				Id: 64,
				Info: &desc.UpdateUserInfo{
					Username: nil,
					Email:    wrapperspb.String("New email"),
				},
			},
			InitMock: func(mock *serviceMocks.UserServiceMock, ctx context.Context, userId int64, dto userService.UpdateUserDTO) {
				mock.UpdateMock.Expect(ctx, userId, dto).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name: "success case - update both name and email",
			req: &desc.UpdateRequest{
				Id: 64,
				Info: &desc.UpdateUserInfo{
					Username: wrapperspb.String("New name"),
					Email:    wrapperspb.String("New email"),
				},
			},
			InitMock: func(mock *serviceMocks.UserServiceMock, ctx context.Context, userId int64, dto userService.UpdateUserDTO) {
				mock.UpdateMock.Expect(ctx, userId, dto).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name: "validation error - id is <= 0",
			req: &desc.UpdateRequest{
				Id: 0,
				Info: &desc.UpdateUserInfo{
					Username: wrapperspb.String("Name"),
				}},
			InitMock: func(mock *serviceMocks.UserServiceMock, ctx context.Context, userId int64, dto userService.UpdateUserDTO) {
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "id is required",
			expectSuccess:  false,
		},
		{
			name: "validation error - info is nil",
			req: &desc.UpdateRequest{
				Id:   52,
				Info: nil},
			InitMock: func(mock *serviceMocks.UserServiceMock, ctx context.Context, userId int64, dto userService.UpdateUserDTO) {
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "info is required",
			expectSuccess:  false,
		},
		{
			name: "validation error - both name and email are nil",
			req: &desc.UpdateRequest{
				Id: 52,
				Info: &desc.UpdateUserInfo{
					Username: nil,
					Email:    nil,
				}},
			InitMock: func(mock *serviceMocks.UserServiceMock, ctx context.Context, userId int64, dto userService.UpdateUserDTO) {
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "at least one field is required",
			expectSuccess:  false,
		},
		{
			name: "service error",
			req: &desc.UpdateRequest{
				Id: 64,
				Info: &desc.UpdateUserInfo{
					Username: wrapperspb.String("New name"),
					Email:    wrapperspb.String("New email"),
				},
			},
			InitMock: func(mock *serviceMocks.UserServiceMock, ctx context.Context, userId int64, dto userService.UpdateUserDTO) {
				mock.UpdateMock.Expect(ctx, userId, dto).Return(errors.New("service error"))
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

			dto := userService.UpdateUserDTO{}
			// Проверяем что Info не nil
			if tt.req.Info != nil {
				if tt.req.Info.Username != nil {
					dto.Username = &tt.req.Info.Username.Value
				}
				if tt.req.Info.Email != nil {
					dto.Email = &tt.req.Info.Email.Value
				}
			}

			tt.InitMock(serviceMock, ctx, tt.req.Id, dto)

			api := user.NewImplementation(serviceMock, nil, &config.Config{}) // jwtManager, cfg

			resp, err := api.Update(ctx, tt.req)

			if tt.expectSuccess {
				require.NoError(t, err)
				require.NotNil(t, resp)
			} else {
				require.Error(t, err)
				require.Nil(t, resp)
				require.Equal(t, tt.expectedCode, status.Code(err))
				require.Contains(t, err.Error(), tt.expectedErrMsg)
			}
		})
	}
}
