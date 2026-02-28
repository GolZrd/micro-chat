package api

import (
	"context"

	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Implementation) PublicChats(ctx context.Context, req *desc.PublicChatsRequest) (*desc.PublicChatsResponse, error) {
	chats, err := s.chatService.PublicChats(ctx, req.Search)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := make([]*desc.PublicChatInfo, 0, len(chats))
	for _, chat := range chats {
		res = append(res, &desc.PublicChatInfo{
			Id:           chat.Id,
			Name:         chat.Name,
			MembersCount: int32(chat.MemberCount),
			CreatorName:  chat.CreatorName,
			CreatedAt:    timestamppb.New(chat.CreatedAt),
		})
	}

	return &desc.PublicChatsResponse{Chats: res}, nil
}
