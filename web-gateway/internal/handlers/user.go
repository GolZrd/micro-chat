package handlers

import (
	"net/http"
	"strconv"

	user_v1 "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"github.com/GolZrd/micro-chat/web-gateway/internal/clients"
	"github.com/GolZrd/micro-chat/web-gateway/internal/utils"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func GetUser(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		resp, err := client.UserClient.Get(ctx, &user_v1.GetRequest{
			Id: userId,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         resp.User.Id,
			"name":       resp.User.Info.Name,
			"email":      resp.User.Info.Email,
			"role":       resp.User.Info.Role.String(),
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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var req struct {
			Name  *string `json:"name"`
			Email *string `json:"email"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updateInfo := user_v1.UpdateUserInfo{}
		if req.Name != nil {
			updateInfo.Name = wrapperspb.String(*req.Name)
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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		_, err = client.UserClient.Delete(ctx, &user_v1.DeleteRequest{
			Id: userId,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	}

}
