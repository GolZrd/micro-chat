package middleware

import "github.com/gin-gonic/gin"

func ExtractTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Извлекаем токен из заголовка authorization
		authHeader := c.GetHeader("authorization")
		if authHeader != "" {
			// Сохраняем токен в контексте для использования в handlers
			c.Set("authorization", authHeader)
		}
		c.Next()
	}
}
