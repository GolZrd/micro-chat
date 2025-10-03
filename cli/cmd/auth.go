package cmd

import (
	"cli/internal/client"
	"cli/internal/token"
	"context"
	"fmt"
	"log"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Команды для аутентификации",
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Регистрация нового пользователя",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")
		email, _ := cmd.Flags().GetString("email")

		fmt.Print("Введите пароль: ")
		password, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print("Подтвердите пароль: ")
		confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println()

		authClient, err := client.NewAuthClient(authServerAddr)
		if err != nil {
			log.Fatal(err)
		}
		defer authClient.Close()

		ctx := context.Background()
		userID, err := authClient.Register(ctx, username, email, string(password), string(confirmPassword))
		if err != nil {
			log.Fatalf("Ошибка регистрации: %v", err)
		}

		fmt.Printf("Пользователь успешно зарегистрирован. ID: %d\n", userID)
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Вход в систему",
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")

		fmt.Print("Введите пароль: ")
		password, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println()

		authClient, err := client.NewAuthClient(authServerAddr)
		if err != nil {
			log.Fatal(err)
		}
		defer authClient.Close()

		ctx := context.Background()
		userInfo, err := authClient.Login(ctx, email, string(password))
		if err != nil {
			log.Fatalf("Ошибка входа: %v", err)
		}

		// Сохраняем токены
		storage := token.NewFileStorage()
		err = storage.SaveUserInfo(userInfo.AccessToken, userInfo.RefreshToken, userInfo.Username)
		if err != nil {
			log.Fatalf("Ошибка сохранения токенов: %v", err)
		}

		fmt.Println("Вход выполнен успешно!")
	},
}

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Обновить access token",
	Run: func(cmd *cobra.Command, args []string) {
		storage := token.NewFileStorage()
		refreshToken, err := storage.GetRefreshToken()
		if err != nil {
			log.Fatalf("Ошибка получения refresh token: %v", err)
		}

		authClient, err := client.NewAuthClient(authServerAddr)
		if err != nil {
			log.Fatal(err)
		}
		defer authClient.Close()

		ctx := context.Background()
		newAccessToken, err := authClient.RefreshAccessToken(ctx, refreshToken)
		if err != nil {
			log.Fatalf("Ошибка обновления токена: %v", err)
		}

		err = storage.SaveAccessToken(newAccessToken)
		if err != nil {
			log.Fatalf("Ошибка сохранения токена: %v", err)
		}

		fmt.Println("Access token успешно обновлен!", newAccessToken)

	},
}

// Команда для показа текущего пользователя
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Показать текущего пользователя",
	Run: func(cmd *cobra.Command, args []string) {
		storage := token.NewFileStorage()
		username, err := storage.GetCurrentUser()
		if err != nil {
			fmt.Println("Не выполнен вход в систему")
			return
		}
		fmt.Printf("Текущий пользователь: %s\n", username)
	},
}

// Команда для переключения между пользователями
var switchCmd = &cobra.Command{
	Use:   "switch [username]",
	Short: "Переключиться на другого пользователя",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		storage := token.NewFileStorage()

		// Если username не указан, показываем список
		if len(args) == 0 {
			users, err := storage.ListUsers()
			if err != nil {
				log.Fatal(err)
			}

			currentUser, _ := storage.GetCurrentUser()

			fmt.Println("Доступные пользователи:")
			for _, user := range users {
				if user == currentUser {
					fmt.Printf("* %s (текущий)\n", user)
				} else {
					fmt.Printf("  %s\n", user)
				}
			}

			fmt.Println("\nИспользуйте: chat-cli auth switch <username>")
			return
		}

		// Переключаемся на указанного пользователя
		username := args[0]
		if err := storage.SwitchUser(username); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Переключено на пользователя: %s\n", username)
	},
}

// Команда для выхода
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Выйти из текущего аккаунта",
	Run: func(cmd *cobra.Command, args []string) {
		storage := token.NewFileStorage()

		currentUser, err := storage.GetCurrentUser()
		if err != nil {
			fmt.Println("Не выполнен вход")
			return
		}

		// Удаляем файл current_user
		err = storage.DeleteUser(currentUser)
		if err != nil {
			log.Fatalf("Ошибка удаления пользователя: %v", err)
		}

		fmt.Printf("Выполнен выход из аккаунта %s\n", currentUser)
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(registerCmd, loginCmd, refreshCmd, whoamiCmd, switchCmd, logoutCmd)

	registerCmd.Flags().StringP("username", "u", "", "Имя пользователя")
	registerCmd.Flags().StringP("email", "e", "", "Email")
	registerCmd.MarkFlagRequired("username")
	registerCmd.MarkFlagRequired("email")

	loginCmd.Flags().StringP("email", "e", "", "Email")
	loginCmd.MarkFlagRequired("email")
}
