package repository

// DTO - в слое репозитория нужен только для записи в базу данных и обновления
// Для чтения у нас также будет использоваться модель User
type CreateUserDTO struct {
	Name     string
	Email    string
	Password string
	Role     string
}

type UpdateUserDTO struct {
	Name  *string
	Email *string
}
