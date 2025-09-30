package client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	chat_v1 "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
)

type ChatClient struct {
	client chat_v1.ChatClient
	conn   *grpc.ClientConn
	token  string
}

func NewChatClient(addr, accessToken string) (*ChatClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("подключение к chat сервису: %w", err)
	}

	return &ChatClient{
		client: chat_v1.NewChatClient(conn),
		conn:   conn,
		token:  accessToken,
	}, nil
}

func (c *ChatClient) Close() error {
	return c.conn.Close()
}

// Добавляем токен в контекст для каждого запроса
func (c *ChatClient) withAuth(ctx context.Context) context.Context {
	fmt.Println("Добавляем токен в контекст", c.token)
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)
}

func (c *ChatClient) CreateChat(ctx context.Context, usernames []string) (int64, error) {
	resp, err := c.client.Create(c.withAuth(ctx), &chat_v1.CreateRequest{
		Usernames: usernames,
	})
	if err != nil {
		return 0, err
	}
	return resp.ChatId, nil
}

func (c *ChatClient) DeleteChat(ctx context.Context, chatID int64) error {
	_, err := c.client.Delete(c.withAuth(ctx), &chat_v1.DeleteRequest{
		Id: chatID,
	})
	return err
}

func (c *ChatClient) SendMessage(ctx context.Context, chatID int64, from_username string, text string) error {
	_, err := c.client.SendMessage(c.withAuth(ctx), &chat_v1.SendMessageRequest{
		ChatId:    chatID,
		From:      from_username,
		Text:      text,
		CreatedAt: timestamppb.Now(),
	})
	return err
}

func (c *ChatClient) ConnectToChat(ctx context.Context, chatID int64) (chat_v1.Chat_ConnectChatClient, error) {
	stream, err := c.client.ConnectChat(c.withAuth(ctx), &chat_v1.ConnectChatRequest{
		ChatId: chatID,
	})
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (c *ChatClient) MyChats(ctx context.Context, username string) ([]*chat_v1.ChatInfo, error) {
	resp, err := c.client.MyChats(c.withAuth(ctx), &chat_v1.MyChatsRequest{
		Username: username,
	})
	if err != nil {
		return nil, err
	}
	return resp.Chats, nil
}
