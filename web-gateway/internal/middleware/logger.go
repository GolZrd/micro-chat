package middleware

import (
	"time"

	"github.com/GolZrd/micro-chat/web-gateway/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Пропускаем статику и health check
		if c.Request.URL.Path == "/health" ||
			c.Request.URL.Path == "/favicon.ico" ||
			len(c.Request.URL.Path) > 7 && c.Request.URL.Path[:7] == "/static" {
			c.Next()
			return
		}

		logger.Info(
			"requst started",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)

		start := time.Now()

		c.Next()

		duration := time.Since(start)

		status := c.Writer.Status()

		if status >= 500 {
			logger.Error(
				"server error",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", status),
				zap.Duration("duration", duration),
			)
		} else if status >= 400 {
			// client error
			logger.Warn(
				"client error",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", status),
				zap.Duration("duration", duration),
			)
		} else {
			// success
			logger.Info(
				"request completed",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", status),
				zap.Duration("duration", duration),
			)
		}
	}
}
