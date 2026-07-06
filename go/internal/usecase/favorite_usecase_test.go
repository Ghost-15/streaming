package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

const (
	favUser  = "user-1"
	favTrack = "track-1"
)

func TestFavoriteUseCase_Add(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		trackID   string
		repoSetup func(*mock.MockFavoriteRepository)
		wantErr   error
	}{
		{
			name:    "success",
			userID:  favUser,
			trackID: favTrack,
			repoSetup: func(r *mock.MockFavoriteRepository) {
				r.AddFn = func(_ context.Context, _, _ string) error { return nil }
			},
		},
		{name: "empty user", userID: "", trackID: favTrack, wantErr: usecase.ErrFavoriteInvalid},
		{name: "empty track", userID: favUser, trackID: "", wantErr: usecase.ErrFavoriteInvalid},
		{
			name:    "repo error",
			userID:  favUser,
			trackID: favTrack,
			repoSetup: func(r *mock.MockFavoriteRepository) {
				r.AddFn = func(_ context.Context, _, _ string) error { return errors.New("db") }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockFavoriteRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			uc := usecase.NewFavoriteUseCase(repo)
			err := uc.Add(context.Background(), tt.userID, tt.trackID)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Add() err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestFavoriteUseCase_Remove(t *testing.T) {
	repo := &mock.MockFavoriteRepository{
		RemoveFn: func(_ context.Context, _, _ string) error { return nil },
	}
	uc := usecase.NewFavoriteUseCase(repo)

	if err := uc.Remove(context.Background(), favUser, favTrack); err != nil {
		t.Fatalf("Remove() err = %v", err)
	}
	if err := uc.Remove(context.Background(), "", favTrack); !errors.Is(err, usecase.ErrFavoriteInvalid) {
		t.Fatalf("Remove() empty user err = %v, want ErrFavoriteInvalid", err)
	}
}

func TestFavoriteUseCase_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mock.MockFavoriteRepository{
			ListByUserFn: func(_ context.Context, _ string) ([]entity.Track, error) {
				return []entity.Track{{ID: favTrack, Title: "Song"}}, nil
			},
		}
		uc := usecase.NewFavoriteUseCase(repo)
		tracks, err := uc.List(context.Background(), favUser)
		if err != nil {
			t.Fatalf("List() err = %v", err)
		}
		if len(tracks) != 1 {
			t.Errorf("List() len = %d, want 1", len(tracks))
		}
	})

	t.Run("empty user", func(t *testing.T) {
		uc := usecase.NewFavoriteUseCase(&mock.MockFavoriteRepository{})
		if _, err := uc.List(context.Background(), ""); !errors.Is(err, usecase.ErrFavoriteInvalid) {
			t.Fatalf("List() err = %v, want ErrFavoriteInvalid", err)
		}
	})
}
