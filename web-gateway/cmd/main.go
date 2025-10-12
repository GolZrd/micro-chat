package main

import (
	"log"
	"os"

	"github.com/GolZrd/micro-chat/web-gateway/internal/clients"
	"github.com/GolZrd/micro-chat/web-gateway/internal/handlers"
	"github.com/GolZrd/micro-chat/web-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Создаем подключения к gRPC сервисам
	authClient, err := clients.NewAuthClient(os.Getenv("AUTH_GRPC_ADDR"))
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %v", err)
	}

	defer authClient.Close()

	chatClient, err := clients.NewChatClient(os.Getenv("CHAT_GRPC_ADDR"))
	if err != nil {
		log.Fatalf("Failed to connect to chat service: %v", err)
	}

	defer chatClient.Close()

	// Создаем HTTP сервер
	r := gin.Default()

	// Middleware для извлечения токена (применяется ко всем запросам)
	r.Use(middleware.ExtractTokenMiddleware())

	// Подключаем статистические файлы
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")
	r.StaticFile("/chat", "./static/chat.html")

	// API endpoints
	api := r.Group("/api")
	{
		// Auth
		api.POST("/register", handlers.Register(authClient))
		api.POST("/login", handlers.Login(authClient))
		api.POST("/refresh", handlers.RefreshAccessToken(authClient))
		api.POST("/refresh-token", handlers.NewRefreshToken(authClient))

		// User
		api.GET("/user/:id", handlers.GetUser(authClient))
		// Chat
		api.POST("/chat/create", handlers.CreateChat(chatClient))
		api.GET("/chat/my", handlers.MyChats(chatClient))
		api.POST("/chat/send", handlers.SendMessage(chatClient))
		api.DELETE("/chat/delete/:id", handlers.DeleteChat(chatClient))
	}

	// WebSocket для чата
	r.GET("/ws/chat/:id", handlers.ConnectChat(chatClient))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)

	r.Run(":" + port)

}
