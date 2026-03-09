package handlers

import (
	"net/http"
	"time"

	"github.com/GolZrd/micro-chat/web-gateway/internal/hub"
	"github.com/GolZrd/micro-chat/web-gateway/internal/logger"
	"github.com/GolZrd/micro-chat/web-gateway/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var notificationsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// NotificationsWS - WebSocket для уведомлений
func NotificationsWS(notificationsHub *hub.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			return
		}

		// Извлекаем user_id из токена
		claims, err := utils.ParseTokenClaims(token)
		if err != nil {
			logger.Error("invalid token for notifications ws", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Upgrade до WebSocket
		conn, err := notificationsUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("notifications ws upgrade failed", zap.Error(err))
			return
		}
		defer conn.Close()

		// Подписываемся на уведомления
		ch := notificationsHub.Subscribe(claims.UserId)
		defer notificationsHub.Unsubscribe(claims.UserId, ch)

		logger.Info("notifications ws connected",
			zap.Int64("user_id", claims.UserId),
			zap.String("username", claims.Username),
		)

		// Горутина для чтения - обнаруживает disconnect
		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					return
				}
			}
		}()

		// Ping ticker
		pingTicker := time.NewTicker(25 * time.Second)
		defer pingTicker.Stop()

		// Основной цикл
		for {
			select {
			case data, ok := <-ch:
				if !ok {
					return
				}

				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					logger.Debug("failed to write message", zap.Error(err))
					return
				}
			case <-pingTicker.C:
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}
}
