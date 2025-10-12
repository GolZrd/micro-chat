package clients

import (
	"fmt"

	auth_v1 "github.com/GolZrd/micro-chat/auth/pkg/auth_v1"
	user_v1 "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	UserClient user_v1.UserAPIClient
	AuthClient auth_v1.AuthAPIClient
	conn       *grpc.ClientConn
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("подключение к auth сервису: %w", err)
	}

	return &AuthClient{
		UserClient: user_v1.NewUserAPIClient(conn),
		AuthClient: auth_v1.NewAuthAPIClient(conn),
		conn:       conn,
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}
