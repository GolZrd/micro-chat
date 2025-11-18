package tests

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/GolZrd/micro-chat/auth/internal/api/access"
	serviceMocks "github.com/GolZrd/micro-chat/auth/internal/service/access/mocks"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	desc "github.com/GolZrd/micro-chat/auth/pkg/access_v1"
)

// Успешный сценарий
func TestCheck_Success(t *testing.T) {
	mc := minimock.NewController(t)

	var (
		endpoint    = fmt.Sprintf("/api/v1/%s", gofakeit.Word())
		accessToken = gofakeit.UUID()

		// контекст с метадатой
		ctx = metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"authorization": "Bearer " + accessToken}),
		)
	)

	serviceMock := serviceMocks.NewAccessServiceMock(mc)
	serviceMock.CheckMock.Expect(ctx, accessToken, endpoint).Return(nil)

	api := access.NewImplementation(serviceMock)

	resp, err := api.Check(ctx, &desc.CheckRequest{EndpointAddress: endpoint})

	require.NoError(t, err)
	assert.NotNil(t, resp)

}

// Тест на пустой эндпоинт
func TestCheck_EmptyEndpoint(t *testing.T) {
	mc := minimock.NewController(t)

	serviceMock := serviceMocks.NewAccessServiceMock(mc)
	api := access.NewImplementation(serviceMock)

	_, err := api.Check(context.Background(), &desc.CheckRequest{EndpointAddress: ""})

	require.Error(t, err)

}

// Тест на отсутствие метадаты
func TestCheck_NoMetadata(t *testing.T) {
	mc := minimock.NewController(t)

	enpoint := "/api/v1/users"

	serviceMock := serviceMocks.NewAccessServiceMock(mc)
	api := access.NewImplementation(serviceMock)

	_, err := api.Check(context.Background(), &desc.CheckRequest{EndpointAddress: enpoint})

	// Assert
	// Проверяем что есть ошибка
	require.Error(t, err)
	// Получаем статус этой ошибки
	st, ok := status.FromError(err)
	// Проверяем что статус есть
	assert.True(t, ok)
	// Сравниваем что ошибка соответствует ожидаемому
	assert.Equal(t, codes.Unauthenticated, st.Code())
	// Проверяем что сообщение соответствует ожидаемому
	assert.Contains(t, st.Message(), "metadata is not provided")
}

func TestCheck_InvalidToken(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    string
		ctx         context.Context
		wantCode    codes.Code
		wantMessage string
	}{
		{
			name:     "no_auth_header",
			endpoint: "/api/v1/users",
			ctx: metadata.NewIncomingContext(
				context.Background(),
				metadata.New(map[string]string{"other": "Value"}),
			),
			wantCode:    codes.Unauthenticated,
			wantMessage: "authorization header is not provided",
		},
		{
			name:     "invalid_token_format",
			endpoint: "/api/v1/users",
			ctx: metadata.NewIncomingContext(
				context.Background(),
				metadata.New(map[string]string{"authorization": "InvalidPrefix token"}),
			),
			wantCode:    codes.InvalidArgument,
			wantMessage: "invalid authorization header format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)

			serviceMock := serviceMocks.NewAccessServiceMock(mc)
			api := access.NewImplementation(serviceMock)

			_, err := api.Check(tt.ctx, &desc.CheckRequest{EndpointAddress: tt.endpoint})

			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.wantCode, st.Code())
			assert.Contains(t, st.Message(), tt.wantMessage)
		})
	}

}

func TestCheck_ServiceError(t *testing.T) {
	mc := minimock.NewController(t)

	var (
		endpoint    = fmt.Sprintf("/api/v1/%s", gofakeit.Word())
		accessToken = gofakeit.UUID()

		// контекст с метадатой
		ctx = metadata.NewIncomingContext(
			context.Background(),
			metadata.New(map[string]string{"authorization": "Bearer " + accessToken}),
		)
		serviceErr = errors.New("access denied")
	)

	serviceMock := serviceMocks.NewAccessServiceMock(mc)
	serviceMock.CheckMock.Expect(ctx, accessToken, endpoint).Return(serviceErr)

	api := access.NewImplementation(serviceMock)

	resp, err := api.Check(ctx, &desc.CheckRequest{EndpointAddress: endpoint})

	// Проверяем что есть ошибка
	require.Error(t, err)
	// Проверяем что resp равен nil
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "access denied")
}
