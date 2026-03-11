package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	user_v1 "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"github.com/GolZrd/micro-chat/web-gateway/internal/clients"
	"github.com/GolZrd/micro-chat/web-gateway/internal/logger"
	"github.com/GolZrd/micro-chat/web-gateway/internal/storage"
	"github.com/GolZrd/micro-chat/web-gateway/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func GetUser(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			logger.Debug("invalid user id", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		resp, err := client.UserClient.Get(ctx, &user_v1.GetRequest{
			Id: userId,
		})
		if err != nil {
			logger.Error("Failed to get user", zap.Int64("user_id", userId), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         resp.User.Id,
			"username":   resp.User.Info.Username,
			"email":      resp.User.Info.Email,
			"role":       resp.User.Info.Role.String(),
			"avatar_url": resp.User.AvatarUrl,
			"bio":        resp.User.Bio,
			"created_at": resp.User.CreatedAt.AsTime(),
			"updated_at": resp.User.UpdatedAt.AsTime(),
		})
	}
}

func UpdateUser(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			logger.Debug("invalid user id", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var req struct {
			Username *string `json:"username"`
			Email    *string `json:"email"`
		}

		if err := c.BindJSON(&req); err != nil {
			logger.Debug("invalid update user request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updateInfo := user_v1.UpdateUserInfo{}
		if req.Username != nil {
			updateInfo.Username = wrapperspb.String(*req.Username)
		}
		if req.Email != nil {
			updateInfo.Email = wrapperspb.String(*req.Email)
		}

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		_, err = client.UserClient.Update(ctx, &user_v1.UpdateRequest{
			Id:   userId,
			Info: &updateInfo,
		})

		if err != nil {
			logger.Error("Failed to update user", zap.Int64("user_id", userId), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	}
}

func DeleteUser(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			logger.Debug("invalid user id", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		_, err = client.UserClient.Delete(ctx, &user_v1.DeleteRequest{
			Id: userId,
		})

		if err != nil {
			logger.Error("Failed to delete user", zap.Int64("user_id", userId), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	}

}

// GET /api/users/search?q=query&limit=20
func SearchUsers(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if len(query) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query too short"})
			return
		}

		limitStr := c.DefaultQuery("limit", "20")
		limit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil || limit <= 0 {
			logger.Debug("invalid limit", zap.Error(err))
			limit = 20
		}

		if limit > 50 {
			limit = 50
		}

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		resp, err := client.UserClient.SearchUser(ctx, &user_v1.SearchUserRequest{
			Query: query,
			Limit: limit,
		})
		if err != nil {
			logger.Error("Failed to search users", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		users := make([]gin.H, 0, len(resp.Users))
		for _, u := range resp.Users {
			users = append(users, gin.H{
				"id":                u.Id,
				"username":          u.Username,
				"friendship_status": u.FriendshipStatus,
				"avatar_url":        u.AvatarUrl,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"users": users,
		})

	}
}

func UploadAvatar(client *clients.AuthClient, storage storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем userId из токена
		token, _ := c.Get("authorization")
		tokenStr, _ := token.(string)
		claims, err := utils.ParseTokenClaims(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Получаем файл
		file, header, err := c.Request.FormFile("avatar")
		if err != nil {
			logger.Debug("invalid avatar request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "avatar file is required"})
			return
		}

		defer file.Close()

		// Проверяем размер (макс 5MB)
		if header.Size > 5*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 5MB)"})
			return
		}

		// Проверяем тип
		contentType := header.Header.Get("Content-Type")
		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/gif":  true,
			"image/webp": true,
		}
		if !allowedTypes[contentType] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only JPEG, PNG, GIF, WebP allowed"})
			return
		}

		// Генерируем имя файла
		ext := ".jpg"
		switch contentType {
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		}
		fileName := fmt.Sprintf("avatars/%d_%d%s", claims.UserId, time.Now().UnixMilli(), ext)

		// Загружаем в MinIO
		fileUrl, err := storage.Upload(c.Request.Context(), fileName, file, header.Size, contentType)
		if err != nil {
			logger.Error("failed to upload avatar", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload"})
			return
		}

		// Обновляем avatar_url в auth-сервисе
		ctx := utils.ContextWithToken(c)
		_, err = client.UserClient.UpdateAvatar(ctx, &user_v1.UpdateAvatarRequest{
			Id:        claims.UserId,
			AvatarUrl: fileUrl,
		})
		if err != nil {
			logger.Error("failed to update avatar", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update avatar"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"avatar_url": fileUrl, "status": "updated"})
	}
}

// UpdateProfile — обновление имени и bio
func UpdateProfile(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Get("authorization")
		tokenStr, _ := token.(string)
		claims, err := utils.ParseTokenClaims(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req struct {
			Username string `json:"username"`
			Bio      string `json:"bio"`
		}

		if err := c.BindJSON(&req); err != nil {
			logger.Debug("invalid update user request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		updateInfo := &user_v1.UpdateUserInfo{}
		if req.Username != "" {
			updateInfo.Username = wrapperspb.String(req.Username)
		}
		if req.Bio != "" {
			updateInfo.Bio = wrapperspb.String(req.Bio)
		}

		_, err = client.UserClient.Update(ctx, &user_v1.UpdateRequest{
			Id:   claims.UserId,
			Info: updateInfo,
		})

		if err != nil {
			logger.Error("Failed to update user", zap.Int64("user_id", claims.UserId), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	}
}
