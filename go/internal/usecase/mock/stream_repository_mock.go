package mock

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

var _ repository.StreamRepository = (*MockStreamRepository)(nil)

type MockStreamRepository struct {
	FindByIDFn           func(ctx context.Context, id string) (*entity.Stream, error)
	ListActiveFn         func(ctx context.Context) ([]entity.Stream, error)
	ListByBroadcasterFn  func(ctx context.Context, broadcasterID string) ([]entity.Stream, error)
	CreateFn             func(ctx context.Context, stream *entity.Stream) error
	ActivateFn           func(ctx context.Context, id, sessionID string) error
	DeactivateFn         func(ctx context.Context, id, sessionID string) (bool, error)
	UpdateStatusFn       func(ctx context.Context, id string, status entity.StreamStatus) error
	DeleteFn             func(ctx context.Context, id string) error
	IncrementListenersFn func(ctx context.Context, id string, delta int) error
}

func (m *MockStreamRepository) FindByID(ctx context.Context, id string) (*entity.Stream, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *MockStreamRepository) ListActive(ctx context.Context) ([]entity.Stream, error) {
	return m.ListActiveFn(ctx)
}

func (m *MockStreamRepository) ListByBroadcaster(ctx context.Context, broadcasterID string) ([]entity.Stream, error) {
	return m.ListByBroadcasterFn(ctx, broadcasterID)
}

func (m *MockStreamRepository) Create(ctx context.Context, stream *entity.Stream) error {
	return m.CreateFn(ctx, stream)
}

func (m *MockStreamRepository) Activate(ctx context.Context, id, sessionID string) error {
	return m.ActivateFn(ctx, id, sessionID)
}

func (m *MockStreamRepository) Deactivate(ctx context.Context, id, sessionID string) (bool, error) {
	return m.DeactivateFn(ctx, id, sessionID)
}

func (m *MockStreamRepository) UpdateStatus(ctx context.Context, id string, status entity.StreamStatus) error {
	return m.UpdateStatusFn(ctx, id, status)
}

func (m *MockStreamRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *MockStreamRepository) IncrementListeners(ctx context.Context, id string, delta int) error {
	return m.IncrementListenersFn(ctx, id, delta)
}
