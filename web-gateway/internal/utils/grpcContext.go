package utils

import (
	"context"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

// ContextWithToken создает gRPC контекст с токеном из Gin контекста
// authorization с маленькой буквы
func ContextWithToken(c *gin.Context) context.Context {
	ctx := context.Background()
	// Извлекаем токен из Gin контекста
	if token, exists := c.Get("authorization"); exists {
		if tokenStr, ok := token.(string); ok {
			// Добавляем токен в gRPC metadata
			md := metadata.New(map[string]string{"authorization": tokenStr})
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
	}

	return ctx
}
