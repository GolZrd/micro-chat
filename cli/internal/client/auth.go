package client

import (
	"context"
	"fmt"

	auth_v1 "github.com/GolZrd/micro-chat/auth/pkg/auth_v1"
	user_v1 "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	userClient user_v1.UserAPIClient
	authClient auth_v1.AuthAPIClient
	conn       *grpc.ClientConn
}

type LoginResponse struct {
	AccessToken  string
	RefreshToken string
	Username     string
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("подключение к auth сервису: %w", err)
	}

	return &AuthClient{
		userClient: user_v1.NewUserAPIClient(conn),
		authClient: auth_v1.NewAuthAPIClient(conn),
		conn:       conn,
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}

func (c *AuthClient) Register(ctx context.Context, username, email, password, passwordConfirm string) (int64, error) {
	userInfo := user_v1.UserInfo{
		Name:            username,
		Email:           email,
		Password:        password,
		PasswordConfirm: passwordConfirm,
		Role:            user_v1.Role_user,
	}

	resp, err := c.userClient.Create(ctx, &user_v1.CreateRequest{
		Info: &userInfo,
	})
	if err != nil {
		return 0, err
	}
	return resp.Id, nil
}

func (c *AuthClient) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	resp, err := c.authClient.Login(ctx, &auth_v1.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	// Для удобства, решил сразу в этом методе на стороне клиента делать запрос на получение access Token
	accessToken, err := c.authClient.GetAccessToken(ctx, &auth_v1.GetAccessTokenRequest{
		RefreshToken: resp.RefreshToken,
	})
	if err != nil {
		return nil, err
	}

	// Получаем информацию по пользователю по его Id
	userInfo, err := c.userClient.Get(ctx, &user_v1.GetRequest{
		//Id: resp.userId,
	})
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken.AccessToken,
		RefreshToken: resp.RefreshToken,
		Username:     userInfo.User.Info.Name,
	}, nil
}

// Метод для обновления access token
func (c *AuthClient) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	resp, err := c.authClient.GetAccessToken(ctx, &auth_v1.GetAccessTokenRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return "", err
	}
	return resp.AccessToken, nil
}

// Метод для обновления refresh token
func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	resp, err := c.authClient.GetRefreshToken(ctx, &auth_v1.GetRefreshTokenRequest{
		OldRefreshToken: refreshToken,
	})
	if err != nil {
		return "", err
	}
	return resp.RefreshToken, nil
}
