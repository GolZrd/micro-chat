package token

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type UserInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username"`
}

type FileStorage struct {
	filePath string
}

func NewFileStorage() *FileStorage {
	homeDir, _ := os.UserHomeDir()
	return &FileStorage{
		filePath: filepath.Join(homeDir, ".chat-cli", "tokens.json"),
	}
}

func (s *FileStorage) SaveUserInfo(accessToken, refreshToken, username string) error {
	userInfo := UserInfo{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Username:     username,
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("создание директории: %w", err)
	}

	data, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("маршалинг токенов: %w", err)
	}

	return os.WriteFile(s.filePath, data, 0600)
}

func (s *FileStorage) SaveAccessToken(accessToken string) error {
	userInfo, err := s.loadUserInfo()
	if err != nil {
		return err
	}
	userInfo.AccessToken = accessToken
	return s.saveUserInfo(userInfo)

}

func (s *FileStorage) GetAccessToken() (string, error) {
	userInfo, err := s.loadUserInfo()
	if err != nil {
		return "", err
	}
	if userInfo.AccessToken == "" {
		return "", fmt.Errorf("access token не найден, выполните вход")
	}
	return userInfo.AccessToken, nil
}

func (s *FileStorage) GetRefreshToken() (string, error) {
	userInfo, err := s.loadUserInfo()
	if err != nil {
		return "", err
	}
	if userInfo.RefreshToken == "" {
		return "", fmt.Errorf("refresh token не найден, выполните вход")
	}
	return userInfo.RefreshToken, nil
}

func (s *FileStorage) loadUserInfo() (*UserInfo, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &UserInfo{}, nil
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
	return os.WriteFile(s.filePath, data, 0600)
}

func (s *FileStorage) GetUsername() (string, error) {
	userInfo, err := s.loadUserInfo()
	if err != nil {
		return "", err
	}

	return userInfo.Username, nil
}
