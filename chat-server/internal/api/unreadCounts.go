package api

import (
	"context"

	"github.com/GolZrd/micro-chat/chat-server/internal/utils"
	desc "github.com/GolZrd/micro-chat/chat-server/pkg/chat_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Implementation) UnreadCounts(ctx context.Context, req *desc.UnreadCountsRequest) (*desc.UnreadCountsResponse, error) {
	userId, err := utils.GetUIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication is required")
	}

	counts, err := s.chatService.UnreadCounts(ctx, userId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get unread counts: %v", err)
	}

	var res []*desc.UnreadCounts
	var total int32
	for chatId, count := range counts {
		res = append(res, &desc.UnreadCounts{
			ChatId: chatId,
			Count:  count,
		})
		total += count
	}

	return &desc.UnreadCountsResponse{
		UnreadCounts: res,
		Total:        total,
	}, nil
}
