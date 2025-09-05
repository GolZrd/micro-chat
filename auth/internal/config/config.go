package config

import (
	"fmt"
	"os"
	"time"
)

// Config — структура для всех настроек
type Config struct {
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	GRPCPort         string
	DB_DSN           string
	RefreshSecretKey string
	AccessSecretKey  string
	RefreshTTL       time.Duration
	AccessTTL        time.Duration
}

// Load загружает конфиг
func Load() (*Config, error) {
	cfg := &Config{
		DBHost:           os.Getenv("DB_HOST"),
		DBPort:           os.Getenv("DB_PORT"),
		DBUser:           os.Getenv("DB_USER"),
		DBPassword:       os.Getenv("DB_PASSWORD"),
		DBName:           os.Getenv("DB_NAME"),
		GRPCPort:         os.Getenv("GRPC_PORT"),
		RefreshSecretKey: os.Getenv("REFRESH_TOKEN_SECRET_KEY"),
		AccessSecretKey:  os.Getenv("ACCESS_TOKEN_SECRET_KEY"),
	}

	accessTTL := os.Getenv("ACCESS_TTL")
	if accessTTL != "" {
		cfg.AccessTTL, _ = time.ParseDuration(accessTTL)
	} else {
		cfg.AccessTTL = 15 * time.Minute
	}

	refreshTTL := os.Getenv("REFRESH_TTL")
	if refreshTTL != "" {
		cfg.RefreshTTL, _ = time.ParseDuration(refreshTTL)
	} else {
		cfg.RefreshTTL = 24 * time.Hour
	}

	cfg.DB_DSN = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPassword)

	return cfg, nil
}
