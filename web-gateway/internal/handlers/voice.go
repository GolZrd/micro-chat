package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	chat_v1 "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"github.com/GolZrd/micro-chat/web-gateway/internal/clients"
	"github.com/GolZrd/micro-chat/web-gateway/internal/logger"
	"github.com/GolZrd/micro-chat/web-gateway/internal/storage"
	"github.com/GolZrd/micro-chat/web-gateway/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const maxVoiceSize = 10 << 20 // 10 MB

func SendVoice(client *clients.ChatClient, storage storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatIdStr := c.PostForm("chat_id")
		chatId, err := strconv.ParseInt(chatIdStr, 10, 64)
		if err != nil {
			logger.Debug("invalid voice request", zap.Error(err), zap.String("chat_id", chatIdStr))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		file, header, err := c.Request.FormFile("voice")
		if err != nil {
			logger.Debug("invalid voice request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "voice file is required"})
			return
		}

		defer file.Close()

		if header.Size > maxVoiceSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 10MB)"})
			return
		}

		// Уникальное имя файла
		filename := fmt.Sprintf("voice/%d_%d.webm", chatId, time.Now().UnixNano())

		// Загружаем в minio
		voiceURL, err := storage.Upload(c.Request.Context(), filename, file, header.Size, "audio/webm")
		if err != nil {
			logger.Error("failed to upload voice", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Длительность
		durationStr := c.PostForm("duration")
		duration, _ := strconv.ParseFloat(durationStr, 64)

		// Создаем контекст с токеном из HTTP заголовка
		ctx := utils.ContextWithToken(c)

		// Сохраняем сообщение в БД
		_, err = client.Client.SendMessage(ctx, &chat_v1.SendMessageRequest{
			ChatId:        chatId,
			Text:          voiceURL,
			Type:          chat_v1.MessageType_MESSAGE_TYPE_VOICE,
			VoiceDuration: float32(duration),
		})
		if err != nil {
			logger.Error("failed to send message", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "sent",
			"voice_url": voiceURL})
	}
}
