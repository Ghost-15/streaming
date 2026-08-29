package mock

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

var _ repository.FavoriteRepository = (*MockFavoriteRepository)(nil)

type MockFavoriteRepository struct {
	AddFn        func(ctx context.Context, userID, streamID string) error
	RemoveFn     func(ctx context.Context, userID, streamID string) error
	ListByUserFn func(ctx context.Context, userID string) ([]entity.Track, error)
}

func (m *MockFavoriteRepository) Add(ctx context.Context, userID, streamID string) error {
	return m.AddFn(ctx, userID, streamID)
}

func (m *MockFavoriteRepository) Remove(ctx context.Context, userID, streamID string) error {
	return m.RemoveFn(ctx, userID, streamID)
}

func (m *MockFavoriteRepository) ListByUser(ctx context.Context, userID string) ([]entity.Track, error) {
	return m.ListByUserFn(ctx, userID)
}
