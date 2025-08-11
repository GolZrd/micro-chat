package converter

import (
	"auth/internal/model"
	desc "auth/pkg/auth_v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Map для конвертации string -> Role enum
var roleFromString = map[string]desc.Role{
	"guest": desc.Role_guest, // 0
	"user":  desc.Role_user,  // 1
	"admin": desc.Role_admin, // 2
}

// Описываем конвертеры для сервисного слоя
func ToUserFromService(user *model.User) *desc.User {
	var updated_at *timestamppb.Timestamp
	if user.UpdatedAt.Valid {
		updated_at = timestamppb.New(user.CreatedAt)
	}

	return &desc.User{
		Id:        user.Id,
		Info:      ToUserInfoFromService(user.Info),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: updated_at,
	}
}

func ToUserInfoFromService(info model.UserInfo) *desc.UserInfo {
	role, ok := roleFromString[info.Role]
	if !ok {
		role = desc.Role_guest
	}

	return &desc.UserInfo{
		Name:     info.Name,
		Email:    info.Email,
		Password: info.Password,
		Role:     role,
	}
}
