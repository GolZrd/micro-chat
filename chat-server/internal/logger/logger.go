package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Создаем глобальный приватный объект
var globalLogger *zap.Logger

// Функция инициализации, будет вызываться при старте приложения
func Init(core zapcore.Core, options ...zap.Option) {
	globalLogger = zap.New(core, options...)
}

// Создаем обертки для логирования
func Debug(msg string, fields ...zap.Field) {
	globalLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	globalLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	globalLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	globalLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	globalLogger.Fatal(msg, fields...)
}

func WithOptions(opts ...zap.Option) *zap.Logger {
	return globalLogger.WithOptions(opts...)
}
