package converter

import (
	"auth/internal/model"
	descUser "auth/pkg/user_v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Map для конвертации string -> Role enum
var roleFromString = map[string]descUser.Role{
	"guest": descUser.Role_guest, // 0
	"user":  descUser.Role_user,  // 1
	"admin": descUser.Role_admin, // 2
}

// Описываем конвертеры для сервисного слоя
func ToUserFromService(user *model.User) *descUser.User {
	var updated_at *timestamppb.Timestamp
	if user.UpdatedAt.Valid {
		updated_at = timestamppb.New(user.CreatedAt)
	}

	return &descUser.User{
		Id:        user.Id,
		Info:      ToUserInfoFromService(user.Info),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: updated_at,
	}
}

func ToUserInfoFromService(info model.UserInfo) *descUser.UserInfo {
	role, ok := roleFromString[info.Role]
	if !ok {
		role = descUser.Role_guest
	}

	return &descUser.UserInfo{
		Name:     info.Name,
		Email:    info.Email,
		Password: info.Password,
		Role:     role,
	}
}
