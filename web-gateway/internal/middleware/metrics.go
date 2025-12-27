package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/GolZrd/micro-chat/web-gateway/internal/metric"
	"github.com/gin-gonic/gin"
)

// Пути которые НЕ измеряем
var skipPrefixes = []string{
	"/static",
	"/favicon",
	"/health",
	"/metrics",
}

// HTML страницы - не API
var htmlPages = map[string]bool{
	"/":     true,
	"/chat": true,
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()

		// Пропускаем статику
		if shouldSkip(path, c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()

		metric.IncHttpRequestsInFlight()
		defer metric.DecHttpRequestsInFlight()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())

		metric.IncHttpRequestTotal(c.Request.Method, path, status)
		metric.ObserveHttpRequestDuration(path, time.Since(start).Seconds())
	}
}

func shouldSkip(templatePath, actualPath string) bool {
	// Пропускаем статику
	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(actualPath, prefix) {
			return true
		}
	}

	// Пропускаем HTML страницы
	if htmlPages[templatePath] {
		return true
	}

	return false
}

func normalizePath(path string) string {
	if path == "" {
		return "not_found"
	}
	return path
}
