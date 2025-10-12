package handlers

import (
	"context"
	"log"
	"net/http"

	auth_v1 "github.com/GolZrd/micro-chat/auth/pkg/auth_v1"
	user_v1 "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"github.com/GolZrd/micro-chat/web-gateway/internal/clients"
	"github.com/gin-gonic/gin"
)

func Register(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name            string `json:"name"`
			Email           string `json:"email"`
			Password        string `json:"password"`
			PasswordConfirm string `json:"password_confirm"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := client.UserClient.Create(context.Background(), &user_v1.CreateRequest{
			Info: &user_v1.UserInfo{
				Name:            req.Name,
				Email:           req.Email,
				Password:        req.Password,
				PasswordConfirm: req.PasswordConfirm,
				Role:            user_v1.Role_user,
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user_id": resp.Id})
	}
}

func Login(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Получаем refresh token
		loginResp, err := client.AuthClient.Login(context.Background(), &auth_v1.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Получаем access token
		accessResp, err := client.AuthClient.GetAccessToken(context.Background(), &auth_v1.GetAccessTokenRequest{
			RefreshToken: loginResp.RefreshToken,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get access token"})
			return
		}

		// Сохраняем refresh token в httpOnly cookie
		c.SetCookie(
			"refresh_token",
			loginResp.RefreshToken,
			24*60*60, // Выставляем время жизни на 24 часа
			"/",
			"",
			false,
			true,
		)

		// Access token отправляем в ответе (клиент сохранит в localStorage)
		c.JSON(http.StatusOK, gin.H{
			"access_token": accessResp.AccessToken,
			"user_id":      loginResp.UserId,
		})
	}
}

// RefreshAccessToken - обновление access token через refresh token из cookie
func RefreshAccessToken(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем refresh token из cookie
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
			return
		}

		log.Printf("Refreshing access token...")

		// Получаем новый access token
		accessResp, err := client.AuthClient.GetAccessToken(context.Background(), &auth_v1.GetAccessTokenRequest{
			RefreshToken: refreshToken,
		})
		if err != nil {
			log.Printf("Failed to refresh access token: %v", err)

			// Удаляем невалидный refresh token cookie
			c.SetCookie(
				"refresh_token",
				"",
				-1, // maxAge -1 = удалить
				"/",
				"",
				false,
				true,
			)

			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}

		log.Printf("Access token refreshed successfully")

		c.JSON(http.StatusOK, gin.H{
			"access_token": accessResp.AccessToken,
		})
	}
}

// Logout - удаление токенов
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Удаляем refresh token cookie
		c.SetCookie(
			"refresh_token",
			"",
			-1, // maxAge -1 = удалить
			"/",
			"",
			false,
			true,
		)

		c.JSON(http.StatusOK, gin.H{"status": "logged out"})
	}
}

// NewRefreshToken - обновление refresh token
func NewRefreshToken(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем старый refresh token из cookie
		oldRefreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
			return
		}

		log.Printf("Updating refresh token...")

		// Получаем новый refresh token
		refreshResp, err := client.AuthClient.GetRefreshToken(context.Background(), &auth_v1.GetRefreshTokenRequest{
			OldRefreshToken: oldRefreshToken,
		})
		if err != nil {
			log.Printf("Failed to update refresh token: %v", err)

			// Удаляем невалидный refresh token cookie
			c.SetCookie(
				"refresh_token",
				"",
				-1, // maxAge -1 = удалить
				"/",
				"",
				false,
				true,
			)

			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}

		log.Printf("Refresh token updated successfully")

		// Сохраняем новый refresh token в cookie
		c.SetCookie(
			"refresh_token",
			refreshResp.RefreshToken,
			24*60*60, // Выставляем время жизни на 24 часа
			"/",
			"",
			false,
			true,
		)

		c.JSON(http.StatusOK, gin.H{"status": "refresh token updated"})
	}
}
