package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// Sentinel errors returned by StreamUseCase, mapped to HTTP codes in the handler.
var (
	ErrStreamNotFound  = errors.New("stream: not found")
	ErrStreamForbidden = errors.New("stream: not the broadcaster")
	ErrStreamInvalid   = errors.New("stream: invalid input")
	ErrStreamEnded     = errors.New("stream: already ended")
)

// StreamUseCase defines the business operations for live streams.
type StreamUseCase interface {
	Start(ctx context.Context, broadcasterID, title, streamURL string) (*entity.Stream, error)
	End(ctx context.Context, streamID, broadcasterID string) error
	ListActive(ctx context.Context) ([]entity.Stream, error)
	Join(ctx context.Context, streamID, userID string) (*entity.Stream, error)
	Leave(ctx context.Context, streamID, userID string) error
}

type streamUseCase struct {
	streamRepo repository.StreamRepository
}

// NewStreamUseCase creates a new StreamUseCase.
func NewStreamUseCase(streamRepo repository.StreamRepository) StreamUseCase {
	return &streamUseCase{streamRepo: streamRepo}
}

// Start creates a new live stream owned by broadcasterID.
func (uc *streamUseCase) Start(ctx context.Context, broadcasterID, title, streamURL string) (*entity.Stream, error) {
	if broadcasterID == "" || strings.TrimSpace(title) == "" {
		return nil, ErrStreamInvalid
	}

	stream := &entity.Stream{
		Title:         strings.TrimSpace(title),
		BroadcasterID: broadcasterID,
		StreamURL:     strings.TrimSpace(streamURL),
		Status:        entity.StreamStatusLive,
	}
	if err := uc.streamRepo.Create(ctx, stream); err != nil {
		return nil, fmt.Errorf("stream: start: %w", err)
	}
	return stream, nil
}

// End stops a stream after verifying broadcasterID is the owner.
func (uc *streamUseCase) End(ctx context.Context, streamID, broadcasterID string) error {
	stream, err := uc.streamRepo.FindByID(ctx, streamID)
	if err != nil {
		return fmt.Errorf("stream: end: %w", err)
	}
	if stream == nil {
		return ErrStreamNotFound
	}
	if stream.BroadcasterID != broadcasterID {
		return ErrStreamForbidden
	}
	if !stream.IsLive() {
		return ErrStreamEnded
	}
	if err := uc.streamRepo.UpdateStatus(ctx, streamID, entity.StreamStatusEnded); err != nil {
		return fmt.Errorf("stream: end: %w", err)
	}
	return nil
}

// ListActive returns all live streams.
func (uc *streamUseCase) ListActive(ctx context.Context) ([]entity.Stream, error) {
	streams, err := uc.streamRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("stream: list active: %w", err)
	}
	return streams, nil
}

// Join registers a listener on a stream and returns the stream (with its audio URL).
func (uc *streamUseCase) Join(ctx context.Context, streamID, userID string) (*entity.Stream, error) {
	if userID == "" {
		return nil, ErrStreamInvalid
	}
	stream, err := uc.streamRepo.FindByID(ctx, streamID)
	if err != nil {
		return nil, fmt.Errorf("stream: join: %w", err)
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}
	if !stream.IsLive() {
		return nil, ErrStreamEnded
	}
	if err := uc.streamRepo.IncrementListeners(ctx, streamID, 1); err != nil {
		return nil, fmt.Errorf("stream: join: %w", err)
	}
	stream.ListenerCount++
	return stream, nil
}

// Leave removes a listener from a stream.
func (uc *streamUseCase) Leave(ctx context.Context, streamID, userID string) error {
	if userID == "" {
		return ErrStreamInvalid
	}
	if err := uc.streamRepo.IncrementListeners(ctx, streamID, -1); err != nil {
		return fmt.Errorf("stream: leave: %w", err)
	}
	return nil
}
