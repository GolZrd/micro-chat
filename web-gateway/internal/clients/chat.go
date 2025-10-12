package clients

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	chat_v1 "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
)

type ChatClient struct {
	Client chat_v1.ChatClient
	conn   *grpc.ClientConn
}

func NewChatClient(addr string) (*ChatClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("подключение к chat сервису: %w", err)
	}

	return &ChatClient{
		Client: chat_v1.NewChatClient(conn),
		conn:   conn,
	}, nil
}

func (c *ChatClient) Close() error {
	return c.conn.Close()
}
