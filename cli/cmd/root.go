package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	authServerAddr string
	chatServerAddr string
)

// Корневая команда
var rootCmd = &cobra.Command{
	Use:   "chat-cli",
	Short: "CLI для работы с микросервисами чата",
	Long:  `CLI утилита для взаимодействия с auth и chat сервисами`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка выполнения команды: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&authServerAddr, "auth-server", "localhost:50051", "Адрес auth сервера")
	rootCmd.PersistentFlags().StringVar(&chatServerAddr, "chat-server", "localhost:50052", "Адрес chat сервера")
}
