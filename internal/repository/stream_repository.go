package repository

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
)

// StreamRepository defines the persistence contract for streams.
// Implemented in internal/infrastructure/supabase/stream_repo.go
type StreamRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Stream, error)
	ListActive(ctx context.Context) ([]entity.Stream, error)
	Create(ctx context.Context, stream *entity.Stream) error
	UpdateStatus(ctx context.Context, id string, status entity.StreamStatus) error
	IncrementListeners(ctx context.Context, id string, delta int) error
}
