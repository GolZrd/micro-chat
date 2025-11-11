package tests

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/GolZrd/micro-chat/auth/internal/api/user"
	"github.com/GolZrd/micro-chat/auth/internal/model"
	serviceMocks "github.com/GolZrd/micro-chat/auth/internal/service/user/mocks"
	desc "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGet(t *testing.T) {

	userData := &model.User{
		Id:        64,
		Info:      model.UserInfo{},
		CreatedAt: time.Now(),
		UpdatedAt: sql.NullTime{},
	}

	tests := []struct {
		name           string
		userId         int64
		InitMock       func(*serviceMocks.UserServiceMock, context.Context, int64)
		expectedUser   *model.User
		expectedCode   codes.Code
		expectedErrMsg string
		expectSuccess  bool
	}{
		{
			name:   "success case",
			userId: int64(64),
			InitMock: func(mock *serviceMocks.UserServiceMock, ctx context.Context, userId int64) {
				mock.GetMock.Expect(ctx, userId).Return(userData, nil)
			},
			expectedUser:  userData,
			expectSuccess: true,
		},
		{
			name:           "validation error - id is <= 0",
			userId:         0,
			InitMock:       func(mock *serviceMocks.UserServiceMock, ctx context.Context, userId int64) {},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "id is required",
			expectSuccess:  false,
		},
		{
			name:   "service error",
			userId: int64(64),
			InitMock: func(mock *serviceMocks.UserServiceMock, ctx context.Context, userId int64) {
				mock.GetMock.Expect(ctx, userId).Return(nil, errors.New("service error"))
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

			tt.InitMock(serviceMock, ctx, tt.userId)

			api := user.NewImplementation(serviceMock)

			resp, err := api.Get(ctx, &desc.GetRequest{Id: tt.userId})

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
