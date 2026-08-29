package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

var ErrFavoriteInvalid = errors.New("favorite: invalid input")

type FavoriteUseCase interface {
	Add(ctx context.Context, userID, streamID string) error
	Remove(ctx context.Context, userID, streamID string) error
	List(ctx context.Context, userID string) ([]entity.Track, error)
}

type favoriteUseCase struct {
	favoriteRepo repository.FavoriteRepository
}

func NewFavoriteUseCase(favoriteRepo repository.FavoriteRepository) FavoriteUseCase {
	return &favoriteUseCase{favoriteRepo: favoriteRepo}
}

func (uc *favoriteUseCase) Add(ctx context.Context, userID, streamID string) error {
	if userID == "" || streamID == "" {
		return ErrFavoriteInvalid
	}
	if err := uc.favoriteRepo.Add(ctx, userID, streamID); err != nil {
		return fmt.Errorf("favorite: add: %w", err)
	}
	return nil
}

func (uc *favoriteUseCase) Remove(ctx context.Context, userID, streamID string) error {
	if userID == "" || streamID == "" {
		return ErrFavoriteInvalid
	}
	if err := uc.favoriteRepo.Remove(ctx, userID, streamID); err != nil {
		return fmt.Errorf("favorite: remove: %w", err)
	}
	return nil
}

func (uc *favoriteUseCase) List(ctx context.Context, userID string) ([]entity.Track, error) {
	if userID == "" {
		return nil, ErrFavoriteInvalid
	}
	tracks, err := uc.favoriteRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("favorite: list: %w", err)
	}
	return tracks, nil
}
