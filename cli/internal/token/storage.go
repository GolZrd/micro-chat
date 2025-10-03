package token

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type UserInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username"`
}

type FileStorage struct {
	baseDir string
}

func NewFileStorage() *FileStorage {
	homeDir, _ := os.UserHomeDir()
	return &FileStorage{
		baseDir: filepath.Join(homeDir, ".chat-cli"),
	}
}

// Сохраняем информацию о текущем активном пользователе
func (s *FileStorage) SetCurrentUser(username string) error {
	currentUserFile := filepath.Join(s.baseDir, "current_user")
	return os.WriteFile(currentUserFile, []byte(username), 0600)
}

func (s *FileStorage) GetCurrentUser() (string, error) {
	currentUserFile := filepath.Join(s.baseDir, "current_user")
	data, err := os.ReadFile(currentUserFile)
	if err != nil {
		return "", fmt.Errorf("не выполнен вход в систему: %w", err)
	}
	return string(data), err
}

// Путь к файлу с токенами для конкретного пользователя
func (s *FileStorage) getUserFilePath(username string) string {
	return filepath.Join(s.baseDir, "users", fmt.Sprintf("%s.json", username))
}

// Теперь сохраняем информацию под конкретным пользователем
func (s *FileStorage) SaveUserInfo(accessToken, refreshToken, username string) error {
	userInfo := UserInfo{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Username:     username,
	}

	// Создаем дирректории
	userDir := filepath.Dir(s.getUserFilePath(username))
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return fmt.Errorf("создание директории: %w", err)
	}

	//Сохраняем данные пользователя
	data, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("маршалинг токенов: %w", err)
	}

	if err := os.WriteFile(s.getUserFilePath(username), data, 0600); err != nil {
		return err
	}

	// Устанавливаем как текущего пользователя
	return s.SetCurrentUser(username)
}

// Сохраняем access токен
func (s *FileStorage) SaveAccessToken(accessToken string) error {
	currentUser, err := s.GetCurrentUser()
	if err != nil {
		return err
	}

	userInfo, err := s.loadUserInfo(currentUser)
	if err != nil {
		return err

	}
	userInfo.AccessToken = accessToken

	return s.saveUserInfo(userInfo)

}

// Получаем access токен текущего пользователя
func (s *FileStorage) GetAccessToken() (string, error) {
	currentUser, err := s.GetCurrentUser()
	if err != nil {
		return "", err
	}

	userInfo, err := s.loadUserInfo(currentUser)
	if err != nil {
		return "", err
	}
	if userInfo.AccessToken == "" {
		return "", fmt.Errorf("access token не найден для пользователя %s, выполните вход", currentUser)
	}
	return userInfo.AccessToken, nil
}

// Получаем refresh token текущего пользователя
func (s *FileStorage) GetRefreshToken() (string, error) {
	currentUser, err := s.GetCurrentUser()
	if err != nil {
		return "", err
	}

	userInfo, err := s.loadUserInfo(currentUser)
	if err != nil {
		return "", err
	}

	if userInfo.RefreshToken == "" {
		return "", fmt.Errorf("refresh token не найден, выполните вход")
	}
	return userInfo.RefreshToken, nil
}

// loadUserInfo загружает информацию конкретного пользователя
func (s *FileStorage) loadUserInfo(username string) (*UserInfo, error) {
	data, err := os.ReadFile(s.getUserFilePath(username))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("пользователь %s не найден", username)
		}
		return nil, fmt.Errorf("чтение файла информации о пользователе: %w", err)
	}

	var userInfo UserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, fmt.Errorf("парсинг информации о пользователе: %w", err)
	}

	return &userInfo, nil
}

func (s *FileStorage) saveUserInfo(userInfo *UserInfo) error {
	data, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("маршалинг информации о пользователе: %w", err)
	}

	return os.WriteFile(s.getUserFilePath(userInfo.Username), data, 0600)
}

// GetUsername возвращает имя текущего пользователя
func (s *FileStorage) GetUsername() (string, error) {
	// userInfo, err := s.loadUserInfo()
	// if err != nil {
	// 	return "", err
	// }

	// return userInfo.Username, nil
	return s.GetCurrentUser()
}

// ListUsers возвращает список пользователей в системе
func (s *FileStorage) ListUsers() ([]string, error) {
	usersDir := filepath.Join(s.baseDir, "users")
	files, err := os.ReadDir(usersDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("чтение списка пользователей: %w", err)
	}

	var users []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			username := strings.TrimSuffix(file.Name(), ".json")
			users = append(users, username)
		}
	}

	return users, nil
}

// SwitchUser переключает активного пользователя
func (s *FileStorage) SwitchUser(username string) error {
	// Сначала проверяем существует ли пользователь
	if _, err := s.loadUserInfo(username); err != nil {
		return fmt.Errorf("пользователь %s не найден, сначала выполните вход", username)
	}

	return s.SetCurrentUser(username)
}

func (s *FileStorage) DeleteUser(username string) error {
	userDir := s.getUserFilePath(username)
	if err := os.RemoveAll(userDir); err != nil {
		return fmt.Errorf("удаление пользователя: %w", err)
	}
	return nil
}
