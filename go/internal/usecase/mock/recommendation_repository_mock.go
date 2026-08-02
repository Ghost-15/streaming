package mock

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

var _ repository.RecommendationRepository = (*MockRecommendationRepository)(nil)

type MockRecommendationRepository struct {
	RecommendStreamsFn func(ctx context.Context, userID string, limit int) ([]entity.Stream, error)
}

func (m *MockRecommendationRepository) RecommendStreams(ctx context.Context, userID string, limit int) ([]entity.Stream, error) {
	return m.RecommendStreamsFn(ctx, userID, limit)
}
