package config

import (
	"fmt"
	"os"
)

// Config — структура для всех настроек
type Config struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	GRPCPort     string
	GRPCAuthPort string
	DB_DSN       string
}

// Load загружает конфиг
func Load() (*Config, error) {
	cfg := &Config{
		DBHost:       os.Getenv("DB_HOST"),
		DBPort:       os.Getenv("DB_PORT"),
		DBUser:       os.Getenv("DB_USER"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
		DBName:       os.Getenv("DB_NAME"),
		GRPCPort:     os.Getenv("GRPC_PORT"),
		GRPCAuthPort: os.Getenv("GRPC_AUTH_PORT"),
	}

	cfg.DB_DSN = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPassword)

	return cfg, nil
}
