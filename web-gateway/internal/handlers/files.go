package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	chat_v1 "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"github.com/GolZrd/micro-chat/web-gateway/internal/clients"
	"github.com/GolZrd/micro-chat/web-gateway/internal/hub"
	"github.com/GolZrd/micro-chat/web-gateway/internal/logger"
	"github.com/GolZrd/micro-chat/web-gateway/internal/storage"
	"github.com/GolZrd/micro-chat/web-gateway/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

var imageExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true,
	".gif": true, ".webp": true, ".svg": true,
}

func SendFile(client *clients.ChatClient, storage storage.Storage, notificationHub *hub.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Chat ID
		chatIdStr := c.PostForm("chat_id")
		chatId, err := strconv.ParseInt(chatIdStr, 10, 64)
		if err != nil || chatId <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
			return
		}

		// Текст
		text := c.PostForm("text")

		// Файл
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
			return
		}
		defer file.Close()

		// Проверка размера (20MB макс)
		if header.Size > 20*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file too large, max 20MB"})
			return
		}

		// Определяем тип
		ext := strings.ToLower(filepath.Ext(header.Filename))
		contentType := header.Header.Get("Content-Type")
		isImage := imageExtensions[ext] || strings.HasPrefix(contentType, "image/")

		// Генерируем путь в MinIO
		folder := "files"
		if isImage {
			folder = "images"
		}
		fileName := fmt.Sprintf("chat/%d/%s/%d_%s",
			chatId, folder, time.Now().UnixMilli(), header.Filename)

		// Загружаем в MinIO
		fileUrl, err := storage.Upload(c.Request.Context(), fileName, file, header.Size, contentType)
		if err != nil {
			logger.Error("failed to upload file", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
			return
		}

		// Определяем тип сообщения
		msgType := chat_v1.MessageType_MESSAGE_TYPE_FILE
		if isImage {
			msgType = chat_v1.MessageType_MESSAGE_TYPE_IMAGE
		}

		// Если текст пустой — ставим имя файла
		if text == "" {
			if isImage {
				text = "Фото"
			} else {
				text = header.Filename
			}
		}

		// Отправляем
		ctx := utils.ContextWithToken(c)
		_, err = client.Client.SendMessage(ctx, &chat_v1.SendMessageRequest{
			ChatId:   chatId,
			Text:     text,
			Type:     msgType,
			FileUrl:  fileUrl,
			FileName: header.Filename,
			FileSize: header.Size,
		})
		if err != nil {
			logger.Error("failed to send message", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send message"})
			return
		}

		// После успешной отправки — уведомления
		token, _ := c.Get("authorization")
		tokenStr, _ := token.(string)
		senderClaims, _ := utils.ParseTokenClaims(tokenStr)

		go func() {
			ctx := utils.ContextWithToken(c)
			sendNotifications(client, notificationHub, ctx, senderClaims, chatId, text, int32(msgType), 0, fileUrl, fileName, header.Size)
		}()

		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"file_url":  fileUrl,
			"file_name": header.Filename,
			"file_size": header.Size,
			"is_image":  isImage,
		})
	}
}
