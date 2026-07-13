package repository

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
)

type ListenHistoryRepository interface {
	Record(ctx context.Context, entry *entity.ListenHistory) error
	ListByUser(ctx context.Context, userID string) ([]entity.ListenHistory, error)
}
