package cmd

import (
	"bufio"
	"cli/internal/client"
	"cli/internal/token"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Команды для работы с чатом",
}

var createChatCmd = &cobra.Command{
	Use:   "create",
	Short: "Создать новый чат",
	Long:  "Создать новый чат с указанными пользователями. Примеры: chat-cli chat create -u user1,user2,user3 или chat-cli chat create --users user1 --users user2 --users user3",
	Run: func(cmd *cobra.Command, args []string) {
		// Получаем список пользователей
		usernames, err := getUsernames(cmd)
		if err != nil {
			log.Fatalf("Ошибка получения списка пользователей: %v", err)
		}

		if len(usernames) == 0 {
			log.Fatal("Необходимо указать хотя бы одного пользователя")
		}

		storage := token.NewFileStorage()
		accessToken, err := storage.GetAccessToken()
		if err != nil {
			log.Fatalf("Ошибка получения токена: %v", err)
		}

		log.Println("Токен успешно получен:", accessToken)

		chatClient, err := client.NewChatClient(chatServerAddr, accessToken)
		if err != nil {
			log.Fatal(err)
		}
		defer chatClient.Close()

		ctx := context.Background()
		chatID, err := chatClient.CreateChat(ctx, usernames)
		if err != nil {
			log.Fatalf("Ошибка создания чата: %v", err)
		}

		fmt.Printf("Чат успешно создан. ID: %d\n", chatID)
	},
}

// Вспомогательная функция для получения списка пользователей из флагов
func getUsernames(cmd *cobra.Command) ([]string, error) {
	// Сначала пробуем получить массив значений
	users, err := cmd.Flags().GetStringSlice("users")
	if err == nil && len(users) > 0 {
		return users, nil
	}

	// Затем пробуем получить строку с разделителями
	usersStr, err := cmd.Flags().GetString("users")
	if err != nil {
		return nil, err
	}

	if usersStr == "" {
		return []string{}, nil
	}

	// Разбиваем строку по запятым и очищаем от пробелов
	usernames := strings.Split(usersStr, ",")
	for i := range usernames {
		usernames[i] = strings.TrimSpace(usernames[i])
	}

	// Фильтруем пустые значения
	filtered := make([]string, 0, len(usernames))
	for _, u := range usernames {
		if u != "" {
			filtered = append(filtered, u)
		}
	}

	return filtered, nil
}

var connectChatCmd = &cobra.Command{
	Use:   "connect",
	Short: "Подключиться к чату",
	Run: func(cmd *cobra.Command, args []string) {
		chatID, _ := cmd.Flags().GetInt64("id")

		storage := token.NewFileStorage()
		accessToken, err := storage.GetAccessToken()
		if err != nil {
			log.Fatalf("Ошибка получения токена: %v", err)
		}

		chatClient, err := client.NewChatClient(chatServerAddr, accessToken)
		if err != nil {
			log.Fatal(err)
		}
		defer chatClient.Close()

		ctx := context.Background()
		stream, err := chatClient.ConnectToChat(ctx, chatID)
		if err != nil {
			log.Fatalf("Ошибка подключения к чату: %v", err)
		}

		fmt.Printf("Подключен к чату %d. Введите сообщения (для выхода введите /exit):\n", chatID)

		// Горутина для получения сообщений
		go func() {
			for {
				msg, err := stream.Recv()
				if err == io.EOF {
					return
				}
				if err != nil {
					log.Printf("Ошибка получения сообщения: %v", err)
					return
				}
				fmt.Printf("\n[%s] %s: %s\n> ", msg.CreatedAt.AsTime().Format("15:04:05"),
					msg.From, msg.Text)
			}
		}()

		// Чтение и отправка сообщений
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("> ")
			scanner.Scan()
			text := scanner.Text()

			if text == "/exit" {
				break
			}

			username, _ := storage.GetUsername()

			if text != "" {
				err := chatClient.SendMessage(ctx, chatID, username, text)
				if err != nil {
					log.Printf("Ошибка отправки сообщения: %v", err)
				}
			}
		}
	},
}

var deleteChatCmd = &cobra.Command{
	Use:   "delete",
	Short: "Удалить чат",
	Run: func(cmd *cobra.Command, args []string) {
		chatID, _ := cmd.Flags().GetInt64("id")

		storage := token.NewFileStorage()
		accessToken, err := storage.GetAccessToken()
		if err != nil {
			log.Fatalf("Ошибка получения токена: %v", err)
		}

		chatClient, err := client.NewChatClient(chatServerAddr, accessToken)
		if err != nil {
			log.Fatal(err)
		}
		defer chatClient.Close()

		ctx := context.Background()
		err = chatClient.DeleteChat(ctx, chatID)
		if err != nil {
			log.Fatalf("Ошибка удаления чата: %v", err)
		}

		fmt.Printf("Чат %d успешно удален\n", chatID)
	},
}

var listChatsCmd = &cobra.Command{
	Use:   "list",
	Short: "Показать мои чаты",
	Run: func(cmd *cobra.Command, args []string) {
		storage := token.NewFileStorage()
		accessToken, err := storage.GetAccessToken()
		if err != nil {
			log.Fatalf("Ошибка получения токена: %v", err)
		}

		chatClient, err := client.NewChatClient(chatServerAddr, accessToken)
		if err != nil {
			log.Fatal(err)
		}
		defer chatClient.Close()
		// Получаем имя текущего пользователя
		username, err := storage.GetUsername()
		if err != nil {
			log.Fatalf("Ошибка получения имени пользователя: %v", err)
		}
		log.Println("my username:", username)

		ctx := context.Background()
		chats, err := chatClient.MyChats(ctx, username)
		if err != nil {
			log.Fatalf("Ошибка получения списка чатов: %v", err)
		}

		if len(chats) == 0 {
			fmt.Println("У вас пока нет чатов")
			return
		}

		fmt.Println("Ваши чаты:")
		fmt.Println("ID\tУчастники\t\t\tСоздан")
		fmt.Println(strings.Repeat("-", 60))

		for _, chat := range chats {
			fmt.Printf("%d\t%s\t%s\n", chat.Id, strings.Join(chat.Usernames, ", "), chat.CreatedAt.AsTime().Format("2006-01-02 15:04:05"))
		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.AddCommand(createChatCmd, connectChatCmd, deleteChatCmd, listChatsCmd)

	// Флаги для create команды - поддерживаем оба варианта
	createChatCmd.Flags().StringP("users", "u", "", "Список пользователей через запятую (например: user1,user2,user3)")
	createChatCmd.Flags().StringSliceP("user", "", []string{}, "Пользователь (можно указать несколько раз)")

	connectChatCmd.Flags().Int64P("id", "i", 0, "ID чата")
	connectChatCmd.MarkFlagRequired("id")

	deleteChatCmd.Flags().Int64P("id", "i", 0, "ID чата")
	deleteChatCmd.MarkFlagRequired("id")
}
