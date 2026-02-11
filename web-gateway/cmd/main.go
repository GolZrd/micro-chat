package main

import (
	"log"
	"net/http"
	"os"

	"github.com/GolZrd/micro-chat/web-gateway/internal/clients"
	"github.com/GolZrd/micro-chat/web-gateway/internal/handlers"
	"github.com/GolZrd/micro-chat/web-gateway/internal/logger"
	"github.com/GolZrd/micro-chat/web-gateway/internal/metric"
	"github.com/GolZrd/micro-chat/web-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	if err := InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.Info("Starting web-gateway")

	// Инициализируем Prometheus метрики
	if err := metric.Init(); err != nil {
		logger.Fatal("Failed to initialize Prometheus metrics", zap.Error(err))
	}

	// Создаем подключения к gRPC сервисам
	authClient, err := clients.NewAuthClient(os.Getenv("AUTH_GRPC_ADDR"))
	if err != nil {
		logger.Fatal("Failed to connect to auth service", zap.Error(err))
	}

	defer func() {
		if err := authClient.Close(); err != nil {
			logger.Error("Failed to close auth client", zap.Error(err))
		}
	}()

	chatClient, err := clients.NewChatClient(os.Getenv("CHAT_GRPC_ADDR"))
	if err != nil {
		logger.Fatal("Failed to connect to chat service", zap.Error(err))
	}

	defer func() {
		if err := chatClient.Close(); err != nil {
			logger.Error("Failed to close chat client", zap.Error(err))
		}
	}()

	logger.Info("Connected to gRPC services")

	// Создаем HTTP сервер
	r := gin.Default()

	// Middleware для извлечения токена (применяется ко всем запросам)
	r.Use(middleware.MetricsMiddleware(), middleware.ExtractTokenMiddleware(), middleware.LoggingMiddleware())

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

		// Search
		api.GET("/users/search", handlers.SearchUsers(authClient))

		// Friends
		api.GET("/friends", handlers.GetFriends(authClient))
		api.GET("/friends/requests", handlers.GetFriendRequests(authClient))
		api.POST("/friends/request", handlers.SendFriendRequest(authClient))
		api.POST("/friends/accept/:id", handlers.AcceptFriendRequest(authClient))
		api.POST("/friends/reject/:id", handlers.RejectFriendRequest(authClient))
		api.DELETE("/friends/:id", handlers.RemoveFriend(authClient))
		// Chat
		api.POST("/chat/create", handlers.CreateChat(chatClient))
		api.GET("/chat/my", handlers.MyChats(chatClient))
		api.POST("/chat/send", handlers.SendMessage(chatClient))
		api.DELETE("/chat/delete/:id", handlers.DeleteChat(chatClient))
	}

	// WebSocket для чата
	r.GET("/ws/chat/:id", handlers.ConnectChat(chatClient))

	// Запускаем HTTP сервер для прометеуса в горутине
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())

		metricsPort := "2114"

		logger.Info("Metrics server starting", zap.String("port", metricsPort))

		if err := http.ListenAndServe(":"+metricsPort, mux); err != nil {
			logger.Fatal("Metrics server failed", zap.Error(err))
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting server", zap.String("port", port))

	r.Run(":" + port)

}

func InitLogger() error {
	// Указываем уровень логирования
	var level zapcore.Level
	if err := level.Set(os.Getenv("WEB_LOG_LVL")); err != nil {
		log.Fatalf("failed to set log level: %v", err)
	}

	atomicLevel := zap.NewAtomicLevelAt(level)

	stdout := zapcore.AddSync(os.Stdout)
	// Настраиваем запись в файл
	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     3, // days
	})

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, atomicLevel),
		zapcore.NewCore(fileEncoder, file, atomicLevel),
	)

	logger.Init(core)

	return nil
}
