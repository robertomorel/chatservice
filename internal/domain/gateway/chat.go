package gateway

import (
	"context"

	"github.com/robertomorel/chatservice/internal/domain/entity"
)

// Interface
type ChatGateway interface {
	CreateChat(ctx context.Context, chat *entity.Chat) error
	FindChatByID(ctx context.Context, chatID string) (*entity.Chat, error)
	SaveChat(ctx context.Context, chat *entity.Chat) error
}
