package usecase

import (
	"context"
	"errors"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// PlaylistUseCase defines the business operations for playlists.
type PlaylistUseCase interface {
	Create(ctx context.Context, ownerID, title string) (*entity.Playlist, error)
	List(ctx context.Context, ownerID string) ([]entity.Playlist, error)
	GetByID(ctx context.Context, id, requesterID string) (*entity.Playlist, error)
	Update(ctx context.Context, id, ownerID, title string) (*entity.Playlist, error)
	Delete(ctx context.Context, id, ownerID string) error
	AddTrack(ctx context.Context, playlistID, ownerID, trackID string) error
	RemoveTrack(ctx context.Context, playlistID, ownerID, trackID string) error
}

type playlistUseCase struct {
	playlistRepo repository.PlaylistRepository
}

// NewPlaylistUseCase creates a new PlaylistUseCase.
func NewPlaylistUseCase(playlistRepo repository.PlaylistRepository) PlaylistUseCase {
	return &playlistUseCase{playlistRepo: playlistRepo}
}

func (uc *playlistUseCase) Create(ctx context.Context, ownerID, title string) (*entity.Playlist, error) {
	// TODO Sprint 2 — US-007
	return nil, errors.New("not implemented")
}

func (uc *playlistUseCase) List(ctx context.Context, ownerID string) ([]entity.Playlist, error) {
	// TODO Sprint 2 — US-007
	return nil, errors.New("not implemented")
}

func (uc *playlistUseCase) GetByID(ctx context.Context, id, requesterID string) (*entity.Playlist, error) {
	// TODO Sprint 2 — US-007
	return nil, errors.New("not implemented")
}

func (uc *playlistUseCase) Update(ctx context.Context, id, ownerID, title string) (*entity.Playlist, error) {
	// TODO Sprint 2 — US-007
	return nil, errors.New("not implemented")
}

func (uc *playlistUseCase) Delete(ctx context.Context, id, ownerID string) error {
	// TODO Sprint 2 — US-007
	return errors.New("not implemented")
}

func (uc *playlistUseCase) AddTrack(ctx context.Context, playlistID, ownerID, trackID string) error {
	// TODO Sprint 2 — US-007
	return errors.New("not implemented")
}

func (uc *playlistUseCase) RemoveTrack(ctx context.Context, playlistID, ownerID, trackID string) error {
	// TODO Sprint 2 — US-007
	return errors.New("not implemented")
}
