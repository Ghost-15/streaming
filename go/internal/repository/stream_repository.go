package repository

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
)

// StreamRepository defines the persistence contract for streams.
// Implemented in internal/infrastructure/supabase/stream_repo.go.
type StreamRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Stream, error)
	ListActive(ctx context.Context) ([]entity.Stream, error)
	ListByBroadcaster(ctx context.Context, broadcasterID string) ([]entity.Stream, error)
	Create(ctx context.Context, stream *entity.Stream) error
	Activate(ctx context.Context, id, sessionID string) error
	Deactivate(ctx context.Context, id, sessionID string) (bool, error)
	UpdateStatus(ctx context.Context, id string, status entity.StreamStatus) error
	Delete(ctx context.Context, id string) error
	IncrementListeners(ctx context.Context, id string, delta int) error
}
