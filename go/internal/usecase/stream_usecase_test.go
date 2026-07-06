package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

func TestStreamUseCase_ListActive(t *testing.T) {
	t.Run("success returns streams", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			ListActiveFn: func(_ context.Context) ([]entity.Stream, error) {
				return []entity.Stream{
					{ID: "s1", Title: "Live A", Status: entity.StreamStatusLive},
					{ID: "s2", Title: "Live B", Status: entity.StreamStatusLive},
				}, nil
			},
		}
		uc := usecase.NewStreamUseCase(repo)

		streams, err := uc.ListActive(context.Background())
		if err != nil {
			t.Fatalf("ListActive() unexpected error = %v", err)
		}
		if len(streams) != 2 {
			t.Errorf("ListActive() len = %d, want 2", len(streams))
		}
	})

	t.Run("repo error is wrapped", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			ListActiveFn: func(_ context.Context) ([]entity.Stream, error) {
				return nil, errors.New("db down")
			},
		}
		uc := usecase.NewStreamUseCase(repo)

		if _, err := uc.ListActive(context.Background()); err == nil {
			t.Error("ListActive() expected error, got nil")
		}
	})
}
