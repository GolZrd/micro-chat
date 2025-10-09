package converter

import (
	"github.com/GolZrd/micro-chat/auth/internal/model"
	modelRepo "github.com/GolZrd/micro-chat/auth/internal/repository/user/model"
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

func ToUserAuthDataFromRepo(user *modelRepo.UserAuthData) *model.UserAuthData {
	return &model.UserAuthData{
		Id:       user.Id,
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
		Role:     user.Role,
	}
}
