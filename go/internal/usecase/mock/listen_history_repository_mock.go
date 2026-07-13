package mock

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

var _ repository.ListenHistoryRepository = (*MockListenHistoryRepository)(nil)

type MockListenHistoryRepository struct {
	RecordFn     func(ctx context.Context, entry *entity.ListenHistory) error
	ListByUserFn func(ctx context.Context, userID string) ([]entity.ListenHistory, error)
}

func (m *MockListenHistoryRepository) Record(ctx context.Context, entry *entity.ListenHistory) error {
	if m.RecordFn == nil {
		return nil
	}
	return m.RecordFn(ctx, entry)
}

func (m *MockListenHistoryRepository) ListByUser(ctx context.Context, userID string) ([]entity.ListenHistory, error) {
	return m.ListByUserFn(ctx, userID)
}
