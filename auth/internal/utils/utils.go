package utils

import "golang.org/x/crypto/bcrypt"

// функция для проверки входного пароля с тем что уже сохранен в базе для этого пользователя
func VerifyPassword(hashedPassword string, candidatePassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(candidatePassword))
	return err == nil
}
