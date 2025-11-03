package middleware

import (
	"github.com/GolZrd/micro-chat/web-gateway/internal/logger"
	"github.com/gin-gonic/gin"
)

func ExtractTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// Будем еще проверять query параметры, так как web socket не передает токен в заголовке
		// Сначала пробуем извлечь токен из заголовка authorization
		authHeader := c.GetHeader("authorization")
		if authHeader != "" {
			logger.Debug("get token from header")
			token = authHeader
		} else if queryToken := c.Query("token"); queryToken != "" {
			logger.Debug("get token from quary")
			token = "Bearer " + queryToken
		}

		if token != "" {
			// Сохраняем токен в контексте для использования в handlers
			c.Set("authorization", token)
		}
		c.Next()
	}
}
