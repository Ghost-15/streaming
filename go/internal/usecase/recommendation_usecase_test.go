package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

func TestRecommendationUseCase_Recommend(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mock.MockRecommendationRepository{
			RecommendStreamsFn: func(_ context.Context, _ string, limit int) ([]entity.Stream, error) {
				if limit <= 0 {
					t.Errorf("limit = %d, want > 0", limit)
				}
				return []entity.Stream{{ID: "s1"}, {ID: "s2"}}, nil
			},
		}
		uc := usecase.NewRecommendationUseCase(repo)
		streams, err := uc.Recommend(context.Background(), "u1")
		if err != nil {
			t.Fatalf("Recommend err = %v", err)
		}
		if len(streams) != 2 {
			t.Errorf("Recommend len = %d, want 2", len(streams))
		}
	})

	t.Run("empty user", func(t *testing.T) {
		uc := usecase.NewRecommendationUseCase(&mock.MockRecommendationRepository{})
		if _, err := uc.Recommend(context.Background(), ""); !errors.Is(err, usecase.ErrRecommendationInvalid) {
			t.Fatalf("Recommend err = %v, want ErrRecommendationInvalid", err)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mock.MockRecommendationRepository{
			RecommendStreamsFn: func(_ context.Context, _ string, _ int) ([]entity.Stream, error) {
				return nil, errors.New("db down")
			},
		}
		uc := usecase.NewRecommendationUseCase(repo)
		if _, err := uc.Recommend(context.Background(), "u1"); err == nil {
			t.Error("Recommend expected error")
		}
	})
}
