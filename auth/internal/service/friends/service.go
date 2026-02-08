package friends

import (
	"context"

	friendsRepository "github.com/GolZrd/micro-chat/auth/internal/repository/friends"
	"github.com/GolZrd/micro-chat/auth/internal/repository/friends/model"
	userRepository "github.com/GolZrd/micro-chat/auth/internal/repository/user"
)

type FriendsService interface {
	SendFriendRequest(ctx context.Context, userId int64, targetUsername string, targetUserId int64) error // Отправить запрос в друзья, передаем свой userId и id друга, которому отправляем запрос
	AcceptFriendRequest(ctx context.Context, requestId int64, userId int64) error                         // Передаются id запроса и id пользователя, которого добавляем в друзья
	RejectFriendRequest(ctx context.Context, requestId int64, userId int64) error
	Friends(ctx context.Context, userid int64) ([]model.Friend, error)
	RemoveFriend(ctx context.Context, userId, friendId int64) error
	FriendRequests(ctx context.Context, userId int64) ([]model.FriendRequest, error)
}

type service struct {
	friendsRepository friendsRepository.FriendsRepository
	userRepository    userRepository.UserRepository
}

func NewService(friendsRepository friendsRepository.FriendsRepository, userRepository userRepository.UserRepository) FriendsService {
	return &service{friendsRepository: friendsRepository, userRepository: userRepository}
}
