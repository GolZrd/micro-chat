package user

import (
	userService "auth/internal/service/user"
	descUser "auth/pkg/user_v1"
	"context"
	"log"
)

func (s *Implementation) Create(ctx context.Context, req *descUser.CreateRequest) (*descUser.CreateRespone, error) {
	log.Println("Пытаемся создать пользователя")
	// proto → service DTO
	input := userService.CreateUserDTO{
		Name:            req.Info.Name,
		Email:           req.Info.Email,
		Password:        req.Info.Password,
		PasswordConfirm: req.Info.PasswordConfirm,
		Role:            req.Info.Role.String(),
	}

	log.Println("Преобразовали входные данные в DTO")

	log.Println("Вызываем метод create у сервисного слоя")
	id, err := s.authService.Create(ctx, input)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return nil, err
	}

	log.Printf("Create user with id: %d", id)

	return &descUser.CreateRespone{
		Id: id,
	}, nil
}
