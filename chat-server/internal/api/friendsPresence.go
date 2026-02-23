package api

import (
	"context"

	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Implementation) FriendsPresence(ctx context.Context, req *desc.FriendsPresenceRequest) (*desc.FriendsPresenceResponse, error) {
	if len(req.UserIds) == 0 {
		return &desc.FriendsPresenceResponse{}, nil
	}

	presences, err := s.chatService.FriendsPresence(ctx, req.UserIds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := make([]*desc.FriendPresence, 0, len(presences))
	for _, p := range presences {
		result = append(result, &desc.FriendPresence{
			UserId:     p.UserId,
			IsOnline:   p.IsOnline,
			LastSeenAt: timestamppb.New(p.LastSeenAt),
		})
	}

	return &desc.FriendsPresenceResponse{
		Friends: result,
	}, nil
}
