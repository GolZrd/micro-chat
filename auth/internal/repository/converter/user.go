package converter

import (
	"auth/internal/model"
	modelRepo "auth/internal/repository/model"
)

// Описываем конвертер, который будет мапить модель репозитория в общую доменную модель
// modelRepo - модель внутри репо слоя, а model - модель внутри доменного слоя
func ToUserFromRepo(user *modelRepo.User) *model.User {

	return &model.User{
		Id:        user.Id,
		Info:      *ToUserInfoFromRepo(user.Info),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func ToUserInfoFromRepo(info modelRepo.UserInfo) *model.UserInfo {

	return &model.UserInfo{
		Name:     info.Name,
		Email:    info.Email,
		Password: info.Password,
		Role:     info.Role,
	}
}
