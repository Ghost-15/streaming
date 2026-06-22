package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// Sentinel errors returned by PlaylistUseCase, mapped to HTTP codes in the handler.
var (
	ErrPlaylistNotFound  = errors.New("playlist: not found")
	ErrPlaylistForbidden = errors.New("playlist: access forbidden")
	ErrPlaylistInvalid   = errors.New("playlist: invalid input")
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

// Create persists a new playlist owned by ownerID and returns it with its generated ID.
func (uc *playlistUseCase) Create(ctx context.Context, ownerID, title string) (*entity.Playlist, error) {
	if ownerID == "" || strings.TrimSpace(title) == "" {
		return nil, ErrPlaylistInvalid
	}

	playlist := &entity.Playlist{
		ID:      uuid.NewString(),
		OwnerID: ownerID,
		Title:   strings.TrimSpace(title),
	}
	if err := uc.playlistRepo.Create(ctx, playlist); err != nil {
		return nil, fmt.Errorf("playlist: create: %w", err)
	}
	return playlist, nil
}

// List returns all playlists owned by ownerID.
func (uc *playlistUseCase) List(ctx context.Context, ownerID string) ([]entity.Playlist, error) {
	if ownerID == "" {
		return nil, ErrPlaylistInvalid
	}
	playlists, err := uc.playlistRepo.ListByOwner(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("playlist: list: %w", err)
	}
	return playlists, nil
}

// GetByID returns the playlist if requesterID is the owner, otherwise 403/404.
func (uc *playlistUseCase) GetByID(ctx context.Context, id, requesterID string) (*entity.Playlist, error) {
	playlist, err := uc.fetchOwned(ctx, id, requesterID)
	if err != nil {
		return nil, err
	}
	return playlist, nil
}

// Update changes the playlist title after ownership check.
func (uc *playlistUseCase) Update(ctx context.Context, id, ownerID, title string) (*entity.Playlist, error) {
	if strings.TrimSpace(title) == "" {
		return nil, ErrPlaylistInvalid
	}

	playlist, err := uc.fetchOwned(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}

	playlist.Title = strings.TrimSpace(title)
	if err := uc.playlistRepo.Update(ctx, playlist); err != nil {
		return nil, fmt.Errorf("playlist: update: %w", err)
	}
	return playlist, nil
}

// Delete removes the playlist after ownership check.
func (uc *playlistUseCase) Delete(ctx context.Context, id, ownerID string) error {
	if _, err := uc.fetchOwned(ctx, id, ownerID); err != nil {
		return err
	}
	if err := uc.playlistRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("playlist: delete: %w", err)
	}
	return nil
}

// AddTrack appends a track to the playlist after ownership check.
func (uc *playlistUseCase) AddTrack(ctx context.Context, playlistID, ownerID, trackID string) error {
	if trackID == "" {
		return ErrPlaylistInvalid
	}
	if _, err := uc.fetchOwned(ctx, playlistID, ownerID); err != nil {
		return err
	}
	track := &entity.Track{
		ID:         trackID,
		PlaylistID: playlistID,
	}
	if err := uc.playlistRepo.AddTrack(ctx, track); err != nil {
		return fmt.Errorf("playlist: add track: %w", err)
	}
	return nil
}

// RemoveTrack removes a track from the playlist after ownership check.
func (uc *playlistUseCase) RemoveTrack(ctx context.Context, playlistID, ownerID, trackID string) error {
	if trackID == "" {
		return ErrPlaylistInvalid
	}
	if _, err := uc.fetchOwned(ctx, playlistID, ownerID); err != nil {
		return err
	}
	if err := uc.playlistRepo.RemoveTrack(ctx, playlistID, trackID); err != nil {
		return fmt.Errorf("playlist: remove track: %w", err)
	}
	return nil
}

// fetchOwned loads a playlist and verifies that ownerID is the owner.
// Returns ErrPlaylistNotFound if missing, ErrPlaylistForbidden on ownership mismatch.
func (uc *playlistUseCase) fetchOwned(ctx context.Context, id, ownerID string) (*entity.Playlist, error) {
	if id == "" || ownerID == "" {
		return nil, ErrPlaylistInvalid
	}
	playlist, err := uc.playlistRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("playlist: lookup: %w", err)
	}
	if playlist == nil {
		return nil, ErrPlaylistNotFound
	}
	if playlist.OwnerID != ownerID {
		return nil, ErrPlaylistForbidden
	}
	return playlist, nil
}
