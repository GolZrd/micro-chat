package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/GolZrd/micro-chat/chat-server/internal/closer"
	"github.com/GolZrd/micro-chat/chat-server/internal/interceptor"
	"github.com/GolZrd/micro-chat/chat-server/internal/logger"
	"github.com/GolZrd/micro-chat/chat-server/internal/metric"
	"github.com/natefinch/lumberjack"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type App struct {
	serviceProvider *serviceProvider
	grpcServer      *grpc.Server
	httpServer      *http.Server
}

// Создаем объект нашей структуры App
func NewApp(ctx context.Context) (*App, error) {
	a := &App{}

	err := a.InitDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) Run() error {
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	go func() {
		if err := a.RunHttpServer(); err != nil {
			log.Fatalf("failed to run http server: %v", err)
		}
	}()

	return a.RunGRPCServer()
}

// инициализируем зависимости
func (a *App) InitDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.InitServiceProvider,
		a.InitLogger,
		a.InitMetrics,
		a.InitGRPCServer,
		a.InitHttpServer,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) InitServiceProvider(_ context.Context) error {
	a.serviceProvider = newServiceProvider()
	return nil
}

// TODO: Правильно ли читать из env или лучше сделать как то из конфига?
func (a *App) InitLogger(_ context.Context) error {
	// Указываем уровень логирования
	var level zapcore.Level
	if err := level.Set(os.Getenv("CHAT_LOG_LVL")); err != nil {
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

func (a *App) InitGRPCServer(ctx context.Context) error {
	// создаем gRPC-сервер c интерцептором
	a.grpcServer = grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(interceptor.MetricsInterceptor, interceptor.LogInterceptor, a.serviceProvider.AuthInterceptor().Unary())),
	)

	reflection.Register(a.grpcServer)

	// Здесь происходит иниициализация зависимостей
	desc.RegisterChatServer(a.grpcServer, a.serviceProvider.ChatImpl(ctx))
	return nil
}

func (a *App) RunGRPCServer() error {
	log.Printf("server listening at %v", a.serviceProvider.cfg.GRPCPort)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", a.serviceProvider.cfg.GRPCPort))
	if err != nil {
		return err
	}
	err = a.grpcServer.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) InitMetrics(ctx context.Context) error {
	return metric.Init(ctx)
}

func (a *App) InitHttpServer(_ context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	a.httpServer = &http.Server{
		Addr:    ":2113",
		Handler: mux,
	}
	return nil
}

func (a *App) RunHttpServer() error {
	if err := a.httpServer.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
