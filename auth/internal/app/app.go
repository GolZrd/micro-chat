package app

import (
	"context"
	"fmt"
	"log"
	"net"

	"auth/internal/closer"
	descAuth "auth/pkg/auth_v1"
	descUser "auth/pkg/user_v1"

	"google.golang.org/grpc"
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

func (a *App) InitGRPCServer(ctx context.Context) error {
	a.grpcServer = grpc.NewServer()
	reflection.Register(a.grpcServer)

	// Здесь происходит иниициализация зависимостей
	descUser.RegisterAuthServer(a.grpcServer, a.serviceProvider.UserImpl(ctx))
	descAuth.RegisterAuthAPIServer(a.grpcServer, a.serviceProvider.AuthImpl(ctx))
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
