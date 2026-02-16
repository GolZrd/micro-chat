package auth

import (
	"context"
	"fmt"

	"github.com/GolZrd/micro-chat/auth/pkg/access_v1"
	"github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserInfo struct {
	Id       int64
	Username string
}

type CheckUsersResult struct {
	FoundUsers    []UserInfo
	NotFoundUsers []string
}

type Client struct {
	conn         *grpc.ClientConn
	accessClient access_v1.AccessAPIClient
	userClient   user_v1.UserAPIClient
}

func NewClient(authServiceAddr string) (*Client, error) {
	conn, err := grpc.NewClient(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	return &Client{
		conn:         conn,
		accessClient: access_v1.NewAccessAPIClient(conn),
		userClient:   user_v1.NewUserAPIClient(conn),
	}, nil
}

func (c *Client) CheckAccess(ctx context.Context, endpoint string) error {
	// Для отладки
	logger.Debug("Check access", zap.String("endpoint", endpoint))

	_, err := c.accessClient.Check(ctx, &access_v1.CheckRequest{
		EndpointAddress: endpoint,
	})
	return err
}

func (c *Client) CheckUsersExists(ctx context.Context, usernames []string) (*CheckUsersResult, error) {
	// Для отладки
	logger.Debug("Check users exists", zap.Strings("usernames", usernames))

	res, err := c.userClient.CheckUsersExists(ctx, &user_v1.CheckUsersExistsRequest{
		Usernames: usernames,
	})
	if err != nil {
		return nil, fmt.Errorf("auth service error: %w", err)
	}

	result := &CheckUsersResult{
		NotFoundUsers: res.NotFoundUsers,
	}

	for _, user := range res.FoundUsers {
		result.FoundUsers = append(result.FoundUsers, UserInfo{
			Id:       user.Id,
			Username: user.Username,
		})
	}

	return result, nil
}

// Close закрывает соединение с auth service
func (c *Client) Close() error {
	return c.conn.Close()
}
