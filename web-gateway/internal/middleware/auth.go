package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
)

func ExtractTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// Будем еще проверять query параметры, так как web socket не передает токен в заголовке
		// Сначала пробуем извлечь токен из заголовка authorization
		authHeader := c.GetHeader("authorization")
		if authHeader != "" {
			log.Println("get token from header")
			token = authHeader
		} else if queryToken := c.Query("token"); queryToken != "" {
			log.Println("get token from quary")
			token = "Bearer " + queryToken
		}

		if token != "" {
			// Сохраняем токен в контексте для использования в handlers
			c.Set("authorization", token)
		}
		c.Next()
	}
}
