package usecase

import (
	"context"
	"errors"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// StreamUseCase defines the business operations for live streams.
type StreamUseCase interface {
	Start(ctx context.Context, broadcasterID, title string) (*entity.Stream, error)
	End(ctx context.Context, streamID, broadcasterID string) error
	ListActive(ctx context.Context) ([]entity.Stream, error)
	Join(ctx context.Context, streamID, userID string) error
	Leave(ctx context.Context, streamID, userID string) error
}

type streamUseCase struct {
	streamRepo repository.StreamRepository
}

// NewStreamUseCase creates a new StreamUseCase.
func NewStreamUseCase(streamRepo repository.StreamRepository) StreamUseCase {
	return &streamUseCase{streamRepo: streamRepo}
}

func (uc *streamUseCase) Start(ctx context.Context, broadcasterID, title string) (*entity.Stream, error) {
	// TODO Sprint 1 — US-003: Create stream, publish to Hub
	return nil, errors.New("not implemented")
}

func (uc *streamUseCase) End(ctx context.Context, streamID, broadcasterID string) error {
	// TODO Sprint 1 — US-003: Update status to ended, close Hub channel
	return errors.New("not implemented")
}

func (uc *streamUseCase) ListActive(ctx context.Context) ([]entity.Stream, error) {
	// TODO Sprint 1 — US-007: Return active streams from repo
	return nil, errors.New("not implemented")
}

func (uc *streamUseCase) Join(ctx context.Context, streamID, userID string) error {
	// TODO Sprint 1 — US-003: Register listener in Hub, increment counter
	return errors.New("not implemented")
}

func (uc *streamUseCase) Leave(ctx context.Context, streamID, userID string) error {
	// TODO Sprint 1 — US-003: Unregister listener, decrement counter
	return errors.New("not implemented")
}
