package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/GolZrd/micro-chat/auth/internal/closer"
	"github.com/GolZrd/micro-chat/auth/internal/interceptor"
	"github.com/GolZrd/micro-chat/auth/internal/logger"
	descAccess "github.com/GolZrd/micro-chat/auth/pkg/access_v1"
	descAuth "github.com/GolZrd/micro-chat/auth/pkg/auth_v1"
	descUser "github.com/GolZrd/micro-chat/auth/pkg/user_v1"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type App struct {
	serviceProvider *serviceProvider
	grpcServer      *grpc.Server
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

	return a.RunGRPCServer()
}

// инициализируем зависимости
func (a *App) InitDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.InitServiceProvider,
		a.InitLogger,
		a.InitGRPCServer,
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
	if err := level.Set(os.Getenv("AUTH_LOG_LVL")); err != nil {
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
	// Добовляем интерцептор
	a.grpcServer = grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.UnaryInterceptor(interceptor.LogInterceptor),
	)

	reflection.Register(a.grpcServer)

	// Здесь происходит иниициализация зависимостей
	descUser.RegisterUserAPIServer(a.grpcServer, a.serviceProvider.UserImpl(ctx))
	descAuth.RegisterAuthAPIServer(a.grpcServer, a.serviceProvider.AuthImpl(ctx))
	descAccess.RegisterAccessAPIServer(a.grpcServer, a.serviceProvider.AccessImpl(ctx))
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
