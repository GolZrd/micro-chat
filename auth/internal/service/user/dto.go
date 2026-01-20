package user

// DTO нужно для преобразования данных в другой формат, то есть, мы будем получать входные данные в одном формате и преобразовывать их в другой, нужный для этого слоя service
// Например, password_confirm нужен только в этом слое, для проверки на совпадение, дальше он нигде не нужен
type CreateUserDTO struct {
	Username        string
	Email           string
	Password        string
	PasswordConfirm string
	Role            string
}

type UpdateUserDTO struct {
	Username *string
	Email    *string
}
