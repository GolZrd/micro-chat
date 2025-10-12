package handlers

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/types/known/timestamppb"

	chat_v1 "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"github.com/GolZrd/micro-chat/web-gateway/internal/clients"
	"github.com/GolZrd/micro-chat/web-gateway/internal/utils"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func CreateChat(client *clients.ChatClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Usernames []string `json:"usernames"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		// Вызываем gRPC - токен автоматически проверится интерцептором chat-server
		resp, err := client.Client.Create(ctx, &chat_v1.CreateRequest{
			Usernames: req.Usernames,
		})
		if err != nil {
			// Интерцептор вернет ошибку если токен невалиден
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

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
			log.Println("failed to get my chats")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"chats": resp.Chats})
	}
}

func SendMessage(client *clients.ChatClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ChatId int64  `json:"chat_id"`
			Text   string `json:"text"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Создаем контекст с токеном - chat-server сам извлечет username
		ctx := utils.ContextWithToken(c)

		log.Printf("Sending message to chat %d: %s", req.ChatId, req.Text)

		// Вызываем gRPC БЕЗ указания From - chat-server извлечет из токена
		_, err := client.Client.SendMessage(ctx, &chat_v1.SendMessageRequest{
			ChatId:    req.ChatId,
			Text:      req.Text,
			CreatedAt: timestamppb.Now(),
		})
		if err != nil {
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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Используем Upgrader для установки соединения WebSocket
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}
		defer ws.Close()

		// Создаем контекст с токеном
		ctx := utils.ContextWithToken(c)

		// Подключаемся к gRPC стриму
		stream, err := client.Client.ConnectChat(ctx, &chat_v1.ConnectChatRequest{
			ChatId: chatId,
		})
		if err != nil {
			log.Println("gRPC stream error:", err)
			return
		}

		// Читаем сообщения из стрима и отправляем их в WebSocket
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println("Stream receive error:", err)
				break
			}

			// Отправляем сообщение в webSocket
			err = ws.WriteJSON(map[string]interface{}{
				"from":      msg.From,
				"text":      msg.Text,
				"createdAt": msg.CreatedAt.AsTime(),
			})
			if err != nil {
				log.Println("WebSocket write error:", err)
				break
			}
		}
	}

}

func DeleteChat(client *clients.ChatClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получам ID чата из URL параметра
		chatId, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
			return
		}

		ctx := utils.ContextWithToken(c)

		log.Printf("Deleting chat %d", chatId)

		// Вызываем grpc метод удаления чата
		_, err = client.Client.Delete(ctx, &chat_v1.DeleteRequest{
			Id: chatId,
		})
		if err != nil {
			log.Printf("Failed to delete chat: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Printf("Chat %d deleted successfully", chatId)

		c.JSON(http.StatusOK, gin.H{"status": "chat deleted successfully"})
	}
}
