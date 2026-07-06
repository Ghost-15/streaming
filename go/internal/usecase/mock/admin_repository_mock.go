package mock

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// Compile-time check: MockAdminRepository implements repository.AdminRepository.
var _ repository.AdminRepository = (*MockAdminRepository)(nil)

// MockAdminRepository is a hand-rolled mock for admin repository.
// Set the function fields to control behavior per test case.
type MockAdminRepository struct {
	ListUsersFn      func(ctx context.Context, page, limit int) ([]entity.User, int, error)
	GetUserFn        func(ctx context.Context, id string) (*entity.User, error)
	UpdateUserRoleFn func(ctx context.Context, id string, role entity.UserRole) error
	SuspendUserFn    func(ctx context.Context, id string, suspend bool) error
	GetStatsFn       func(ctx context.Context) (*entity.AdminStats, error)
}

func (m *MockAdminRepository) ListUsers(ctx context.Context, page, limit int) ([]entity.User, int, error) {
	return m.ListUsersFn(ctx, page, limit)
}

func (m *MockAdminRepository) GetUser(ctx context.Context, id string) (*entity.User, error) {
	return m.GetUserFn(ctx, id)
}

func (m *MockAdminRepository) UpdateUserRole(ctx context.Context, id string, role entity.UserRole) error {
	return m.UpdateUserRoleFn(ctx, id, role)
}

func (m *MockAdminRepository) SuspendUser(ctx context.Context, id string, suspend bool) error {
	return m.SuspendUserFn(ctx, id, suspend)
}

func (m *MockAdminRepository) GetStats(ctx context.Context) (*entity.AdminStats, error) {
	return m.GetStatsFn(ctx)
}
