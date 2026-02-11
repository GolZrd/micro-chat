package handlers

import (
	"net/http"
	"strconv"

	"github.com/GolZrd/micro-chat/auth/pkg/friends_v1"
	"github.com/GolZrd/micro-chat/web-gateway/internal/clients"
	"github.com/GolZrd/micro-chat/web-gateway/internal/logger"
	"github.com/GolZrd/micro-chat/web-gateway/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GET /api/friends
func GetFriends(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		resp, err := client.FriendsClient.GetFriends(ctx, &friends_v1.GetFriendsRequest{})
		if err != nil {
			logger.Error("Failed to get friends", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		friends := make([]gin.H, 0, len(resp.Friends))
		for _, friend := range resp.Friends {
			friends = append(friends, gin.H{
				"id":       friend.Id,
				"user_id":  friend.UserId,
				"username": friend.Username,
			})
		}

		c.JSON(http.StatusOK, gin.H{"friends": friends})
	}
}

// GET /api/friends/requests
func GetFriendRequests(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		resp, err := client.FriendsClient.GetFriendRequests(ctx, &friends_v1.GetFriendRequestsRequest{})
		if err != nil {
			logger.Error("Failed to get friend requests", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		requests := make([]gin.H, 0, len(resp.Requests))
		for _, req := range resp.Requests {
			requests = append(requests, gin.H{
				"id":            req.Id,
				"from_user_id":  req.FromUserId,
				"from_username": req.FromUsername,
				"created_at":    req.CreatedAt.AsTime(),
			})
		}

		c.JSON(http.StatusOK, gin.H{"requests": requests})
	}
}

// POST /api/friends/request
func SendFriendRequest(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			UserId   int64  `json:"user_id"`
		}

		if err := c.BindJSON(&req); err != nil {
			logger.Debug("invalid send friend request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		_, err := client.FriendsClient.SendFriendRequest(ctx, &friends_v1.SendFriendRequestRequest{
			Username: req.Username,
			UserId:   req.UserId,
		})
		if err != nil {
			logger.Error("Failed to send friend request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "sent"})
	}
}

// POST /api/friends/accept/:id
func AcceptFriendRequest(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			logger.Debug("invalid friend request id", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		_, err = client.FriendsClient.AcceptFriendRequest(ctx, &friends_v1.AcceptFriendRequestRequest{
			RequestId: requestId,
		})
		if err != nil {
			logger.Error("Failed to accept friend request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

// POST /api/friends/reject/:id
func RejectFriendRequest(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			logger.Debug("invalid friend request id", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		_, err = client.FriendsClient.RejectFriendRequest(ctx, &friends_v1.RejectFriendRequestRequest{
			RequestId: requestId,
		})
		if err != nil {
			logger.Error("Failed to accept friend request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

// DELETE /api/friends/:id
func RemoveFriend(client *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		friendId, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			logger.Debug("invalid friend id", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		_, err = client.FriendsClient.RemoveFriend(ctx, &friends_v1.RemoveFriendRequest{
			FriendId: friendId,
		})
		if err != nil {
			logger.Error("Failed to remove friend", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
