package access

import (
	"context"
	"fmt"

	access_v1 "github.com/GolZrd/micro-chat/auth/pkg/access_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	serverAddress = "localhost:50051"
	authPrefix    = "Bearer "
)

// Клиент для взаимодействия с сервисом авторизации для проверки доступа к ресурсу
type Client struct {
	api access_v1.AccessAPIClient
}

func NewClient(authServiceAddr string) (*Client, error) {
	conn, err := grpc.NewClient(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	return &Client{
		api: access_v1.NewAccessAPIClient(conn),
	}, nil
}

func (c *Client) CheckAccess(ctx context.Context, endpoint string) error {
	_, err := c.api.Check(ctx, &access_v1.CheckRequest{
		EndpointAddress: endpoint,
	})
	return err
}
