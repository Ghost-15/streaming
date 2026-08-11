package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
	"github.com/Ghost-15/streaming/internal/repository"
)

var (
	ErrStreamInvalid        = errors.New("stream: invalid input")
	ErrStreamNotFound       = errors.New("stream: not found")
	ErrStreamForbidden      = errors.New("stream: access forbidden")
	ErrStreamSessionExpired = errors.New("stream: broadcast session expired")
)

// StreamUseCase defines the business operations for live streams.
type StreamUseCase interface {
	Start(ctx context.Context, broadcasterID, title string) (*entity.Stream, error)
	Restart(ctx context.Context, streamID, broadcasterID string) (*entity.Stream, error)
	End(ctx context.Context, streamID, broadcasterID, sessionID string) error
	Delete(ctx context.Context, streamID, broadcasterID string) error
	ListActive(ctx context.Context) ([]entity.Stream, error)
	ListOwned(ctx context.Context, broadcasterID string) ([]entity.Stream, error)
	Join(ctx context.Context, streamID, userID string) error
	Leave(ctx context.Context, streamID, userID string) error
	CanBroadcast(ctx context.Context, streamID, broadcasterID, sessionID string) error
}

func (uc *streamUseCase) CanBroadcast(ctx context.Context, streamID, broadcasterID, sessionID string) error {
	if sessionID == "" {
		return ErrStreamInvalid
	}
	stream, err := uc.fetchOwned(ctx, streamID, broadcasterID)
	if err != nil {
		return err
	}
	if !stream.IsLive() || stream.ActiveSessionID == nil || *stream.ActiveSessionID != sessionID {
		return ErrStreamSessionExpired
	}
	return nil
}

type streamUseCase struct {
	streamRepo  repository.StreamRepository
	historyRepo repository.ListenHistoryRepository
}

// NewStreamUseCase creates a new StreamUseCase.
// historyRepo may be nil: listen history recording is best-effort.
func NewStreamUseCase(streamRepo repository.StreamRepository, historyRepo repository.ListenHistoryRepository) StreamUseCase {
	return &streamUseCase{streamRepo: streamRepo, historyRepo: historyRepo}
}

func (uc *streamUseCase) Start(ctx context.Context, broadcasterID, title string) (*entity.Stream, error) {
	if broadcasterID == "" || strings.TrimSpace(title) == "" {
		return nil, ErrStreamInvalid
	}

	stream := &entity.Stream{
		Title:         strings.TrimSpace(title),
		BroadcasterID: broadcasterID,
		Status:        entity.StreamStatusLive,
		StartedAt:     time.Now(),
	}
	sessionID := uuid.NewString()
	stream.ActiveSessionID = &sessionID
	if err := uc.streamRepo.Create(ctx, stream); err != nil {
		return nil, fmt.Errorf("stream: start: %w", err)
	}
	telemetry.IncrementActiveStreams()
	telemetry.RecordStreamStart()
	return stream, nil
}

func (uc *streamUseCase) Restart(ctx context.Context, streamID, broadcasterID string) (*entity.Stream, error) {
	stream, err := uc.fetchOwned(ctx, streamID, broadcasterID)
	if err != nil {
		return nil, err
	}
	wasLive := stream.IsLive()
	sessionID := uuid.NewString()
	if err := uc.streamRepo.Activate(ctx, stream.ID, sessionID); err != nil {
		return nil, fmt.Errorf("stream: restart: %w", err)
	}
	stream.Status = entity.StreamStatusLive
	stream.ActiveSessionID = &sessionID
	stream.StartedAt = time.Now()
	stream.EndedAt = nil
	stream.ListenerCount = 0
	if !wasLive {
		telemetry.IncrementActiveStreams()
		telemetry.RecordStreamStart()
	}
	return stream, nil
}

func (uc *streamUseCase) End(ctx context.Context, streamID, broadcasterID, sessionID string) error {
	if sessionID == "" {
		return ErrStreamInvalid
	}
	stream, err := uc.fetchOwned(ctx, streamID, broadcasterID)
	if err != nil {
		return err
	}
	if !stream.IsLive() {
		return nil
	}
	if stream.ActiveSessionID == nil || *stream.ActiveSessionID != sessionID {
		return ErrStreamSessionExpired
	}
	deactivated, err := uc.streamRepo.Deactivate(ctx, stream.ID, sessionID)
	if err != nil {
		return fmt.Errorf("stream: end: %w", err)
	}
	if !deactivated {
		return ErrStreamSessionExpired
	}
	telemetry.DecrementActiveStreams()
	return nil
}

func (uc *streamUseCase) Delete(ctx context.Context, streamID, broadcasterID string) error {
	stream, err := uc.fetchOwned(ctx, streamID, broadcasterID)
	if err != nil {
		return err
	}
	if err := uc.streamRepo.Delete(ctx, stream.ID); err != nil {
		return fmt.Errorf("stream: delete: %w", err)
	}
	if stream.IsLive() {
		telemetry.DecrementActiveStreams()
	}
	return nil
}

func (uc *streamUseCase) ListActive(ctx context.Context) ([]entity.Stream, error) {
	streams, err := uc.streamRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("stream: list active: %w", err)
	}
	return streams, nil
}

func (uc *streamUseCase) ListOwned(ctx context.Context, broadcasterID string) ([]entity.Stream, error) {
	if broadcasterID == "" {
		return nil, ErrStreamInvalid
	}
	streams, err := uc.streamRepo.ListByBroadcaster(ctx, broadcasterID)
	if err != nil {
		return nil, fmt.Errorf("stream: list owned: %w", err)
	}
	return streams, nil
}

func (uc *streamUseCase) Join(ctx context.Context, streamID, userID string) error {
	if streamID == "" || userID == "" {
		return ErrStreamInvalid
	}
	stream, err := uc.streamRepo.FindByID(ctx, streamID)
	if err != nil {
		return fmt.Errorf("stream: join lookup: %w", err)
	}
	if stream == nil || !stream.IsLive() {
		return ErrStreamNotFound
	}
	if err := uc.streamRepo.IncrementListeners(ctx, streamID, 1); err != nil {
		return fmt.Errorf("stream: join: %w", err)
	}
	uc.recordStreamEvent(ctx, streamID, userID, entity.ListenEventJoin)
	return nil
}

// recordStreamEvent persists stream listen-history events best-effort.
func (uc *streamUseCase) recordStreamEvent(ctx context.Context, streamID, userID string, event entity.ListenEvent) {
	if uc.historyRepo == nil {
		return
	}
	sID := streamID
	_ = uc.historyRepo.Record(ctx, &entity.ListenHistory{
		UserID:     userID,
		StreamID:   &sID,
		Event:      event,
		ListenedAt: time.Now(),
	})
}

func (uc *streamUseCase) Leave(ctx context.Context, streamID, userID string) error {
	if streamID == "" || userID == "" {
		return ErrStreamInvalid
	}
	if err := uc.streamRepo.IncrementListeners(ctx, streamID, -1); err != nil {
		return fmt.Errorf("stream: leave: %w", err)
	}
	uc.recordStreamEvent(ctx, streamID, userID, entity.ListenEventLeave)
	return nil
}

func (uc *streamUseCase) fetchOwned(ctx context.Context, streamID, broadcasterID string) (*entity.Stream, error) {
	if streamID == "" || broadcasterID == "" {
		return nil, ErrStreamInvalid
	}
	stream, err := uc.streamRepo.FindByID(ctx, streamID)
	if err != nil {
		return nil, fmt.Errorf("stream: lookup: %w", err)
	}
	if stream == nil {
		return nil, ErrStreamNotFound
	}
	if stream.BroadcasterID != broadcasterID {
		return nil, ErrStreamForbidden
	}
	return stream, nil
}
