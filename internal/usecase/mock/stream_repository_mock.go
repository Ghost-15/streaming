package mock

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

var _ repository.StreamRepository = (*MockStreamRepository)(nil)

type MockStreamRepository struct {
	FindByIDFn          func(ctx context.Context, id string) (*entity.Stream, error)
	ListActiveFn        func(ctx context.Context) ([]entity.Stream, error)
	CreateFn            func(ctx context.Context, stream *entity.Stream) error
	UpdateStatusFn      func(ctx context.Context, id string, status entity.StreamStatus) error
	IncrementListenersFn func(ctx context.Context, id string, delta int) error
}

func (m *MockStreamRepository) FindByID(ctx context.Context, id string) (*entity.Stream, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *MockStreamRepository) ListActive(ctx context.Context) ([]entity.Stream, error) {
	return m.ListActiveFn(ctx)
}

func (m *MockStreamRepository) Create(ctx context.Context, stream *entity.Stream) error {
	return m.CreateFn(ctx, stream)
}

func (m *MockStreamRepository) UpdateStatus(ctx context.Context, id string, status entity.StreamStatus) error {
	return m.UpdateStatusFn(ctx, id, status)
}

func (m *MockStreamRepository) IncrementListeners(ctx context.Context, id string, delta int) error {
	return m.IncrementListenersFn(ctx, id, delta)
}
