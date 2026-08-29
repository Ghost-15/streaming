package repository

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
)

type FavoriteRepository interface {
	Add(ctx context.Context, userID, streamID string) error
	Remove(ctx context.Context, userID, streamID string) error
	ListByUser(ctx context.Context, userID string) ([]entity.Track, error)
}
