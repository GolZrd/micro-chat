package handlers

import (
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

func ProxyFiles() gin.HandlerFunc {
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "minio:9000"
	}

	minioURL, _ := url.Parse("http://" + minioEndpoint)
	proxy := httputil.NewSingleHostReverseProxy(minioURL)

	return func(c *gin.Context) {
		c.Request.URL.Path = c.Param("filepath")
		c.Request.Host = minioURL.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
