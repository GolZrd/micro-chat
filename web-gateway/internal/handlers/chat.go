package handlers

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

	chat_v1 "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"github.com/GolZrd/micro-chat/web-gateway/internal/clients"
	"github.com/GolZrd/micro-chat/web-gateway/internal/logger"
	"github.com/GolZrd/micro-chat/web-gateway/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func CreateChat(client *clients.ChatClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name      string   `json:"name"`
			Usernames []string `json:"usernames"`
		}

		if err := c.BindJSON(&req); err != nil {
			logger.Debug("invalid create chat request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.Info("create chat attempt",
			zap.Strings("usernames", req.Usernames),
		)

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		// Вызываем gRPC - токен автоматически проверится интерцептором chat-server
		resp, err := client.Client.Create(ctx, &chat_v1.CreateRequest{
			Name:      req.Name,
			Usernames: req.Usernames,
		})
		if err != nil {
			logger.Error("failed to create chat", zap.Error(err))
			handleChatError(c, err)
			return
		}

		logger.Info("chat created", zap.Int64("chat_id", resp.ChatId))

		c.JSON(http.StatusOK, gin.H{"chat_id": resp.ChatId})
	}
}

// MyChats возвращает список чатов пользователя
func MyChats(client *clients.ChatClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		resp, err := client.Client.MyChats(ctx, &chat_v1.MyChatsRequest{})
		if err != nil {
			logger.Error("failed to get my chats")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		chats := make([]gin.H, 0, len(resp.Chats))
		for _, ch := range resp.Chats {
			chats = append(chats, gin.H{
				"id":         ch.Id,
				"name":       ch.Name,
				"is_direct":  ch.IsDirect,
				"usernames":  ch.Usernames,
				"created_at": ch.CreatedAt.AsTime(),
			})
		}

		logger.Info("got user chats", zap.Int("count", len(chats)))

		c.JSON(http.StatusOK, gin.H{"chats": chats})
	}
}

func SendMessage(client *clients.ChatClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ChatId int64  `json:"chat_id"`
			Text   string `json:"text"`
		}

		if err := c.BindJSON(&req); err != nil {
			logger.Debug("invalid send message request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Создаем контекст с токеном - chat-server сам извлечет username
		ctx := utils.ContextWithToken(c)

		// Вызываем gRPC БЕЗ указания From - chat-server извлечет из токена
		_, err := client.Client.SendMessage(ctx, &chat_v1.SendMessageRequest{
			ChatId: req.ChatId,
			Text:   req.Text,
		})
		if err != nil {
			logger.Error("failed to send message", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "sent"})
	}
}

func ConnectChat(client *clients.ChatClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatId, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			logger.Warn("invalid chat id", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Используем Upgrader для установки соединения WebSocket
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("WebSocket upgrade failed", zap.Error(err))
			return
		}
		defer ws.Close()

		// Создаем контекст с Отменой через контекст с токеном
		ctx, cancel := context.WithCancel(utils.ContextWithToken(c))
		defer cancel() // Отменяем контекст

		// Подключаемся к gRPC стриму
		stream, err := client.Client.ConnectChat(ctx, &chat_v1.ConnectChatRequest{
			ChatId: chatId,
		})
		if err != nil {
			logger.Error("failed to connect to chat", zap.Error(err))
			return
		}

		logger.Info("connected to chat", zap.Int64("chat_id", chatId))

		// Канал для завершения webSocket соединения
		done := make(chan struct{})

		// Делаем горутину, которая отслеживает закрытие webSocket соединения
		go func() {
			defer close(done)

			// ждем закрытия webSocket
			for {
				_, _, err := ws.ReadMessage()
				if err != nil {
					logger.Debug("Websocket closed", zap.Int64("chat_id", chatId))
					cancel() // Отменяем контекст
					return
				}
			}
		}()

		// Читаем сообщения из стрима и отправляем их в WebSocket
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF || ctx.Err() != nil {
					break
				}
				logger.Error("failed to receive message", zap.Error(err))
				break
			}

			wsMsg := convertToWebSocketMessage(msg)

			// Отправляем сообщение в webSocket
			err = ws.WriteJSON(wsMsg)
			if err != nil {
				logger.Error("WebSocket write error", zap.Error(err))
				break
			}
		}

		// Ожидаем завершения webSocket соединения
		<-done
	}

}

func DeleteChat(client *clients.ChatClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получам ID чата из URL параметра
		chatId, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			logger.Warn("invalid chat id", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
			return
		}

		ctx := utils.ContextWithToken(c)

		// Вызываем grpc метод удаления чата
		_, err = client.Client.Delete(ctx, &chat_v1.DeleteRequest{
			Id: chatId,
		})
		if err != nil {
			logger.Error("Failed to delete chat", zap.Int64("chat_id", chatId), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Chat deleted successfully", zap.Int64("chat_id", chatId))

		c.JSON(http.StatusOK, gin.H{"status": "chat deleted successfully"})
	}
}

func GetOrCreateDirectChat(client *clients.ChatClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserId   int64  `json:"user_id"`
			Username string `json:"username"`
		}

		if err := c.BindJSON(&req); err != nil {
			logger.Debug("invalid get or create direct chat request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := utils.ContextWithToken(c)

		resp, err := client.Client.GetOrCreateDirectChat(ctx, &chat_v1.GetOrCreateDirectChatRequest{
			UserId:   req.UserId,
			Username: req.Username,
		})
		if err != nil {
			logger.Error("failed to get or create direct chat", zap.Error(err))
			handleChatError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"chat_id": resp.ChatId, "created": resp.Created})
	}
}

func handleChatError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Внутренняя ошибка",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	msg := st.Message()
	switch st.Code() {
	case codes.NotFound:
		if strings.HasPrefix(msg, "USERS_NOT_FOUND:") {
			usersList := strings.TrimPrefix(msg, "USERS_NOT_FOUND:")
			users := strings.Split(usersList, ",")

			c.JSON(http.StatusNotFound, gin.H{
				"error":           "Пользователи не найдены",
				"code":            "USERS_NOT_FOUND",
				"not_found_users": users,
			})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{
			"error": msg,
		})
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": msg,
			"code":  "INVALID_ARGUMENT",
		})
	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Требуется авторизация",
			"code":  "UNAUTHENTICATED",
		})
	case codes.PermissionDenied:
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Доступ запрещен",
			"code":  "PERMISSION_DENIED",
		})
	default:
		logger.Error("grpc error", zap.String("code", st.Code().String()), zap.String("msg", msg))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Внутренняя ошибка",
			"code":  "INTERNAL_ERROR",
		})
	}
}

func convertToWebSocketMessage(msg *chat_v1.Message) map[string]interface{} {
	switch msg.Type {
	case chat_v1.MessageType_MESSAGE_TYPE_ONLINE_USERS:
		onlineUsers := make([]map[string]interface{}, 0, len(msg.OnlineUsers))
		for _, user := range msg.OnlineUsers {
			onlineUsers = append(onlineUsers, map[string]interface{}{
				"userId":   user.UserId,
				"username": user.Username,
			})
		}

		return map[string]interface{}{
			"type":        "online_users",
			"onlineUsers": onlineUsers,
			"onlineCount": len(msg.OnlineUsers),
		}
	default:
		return map[string]interface{}{
			"type":    "message",
			"from":    msg.From,
			"text":    msg.Text,
			"sent_at": msg.CreatedAt.AsTime(),
		}
	}

}
