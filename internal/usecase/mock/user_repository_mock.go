// Package mock provides test doubles for repository interfaces.
// Used exclusively in usecase tests — never in production code.
package mock

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// Compile-time check: MockUserRepository implements repository.UserRepository.
var _ repository.UserRepository = (*MockUserRepository)(nil)

// MockUserRepository is a hand-rolled mock — no external mock framework needed.
// Set the function fields to control behavior per test case.
type MockUserRepository struct {
	FindByEmailFn func(ctx context.Context, email string) (*entity.User, error)
	FindByIDFn    func(ctx context.Context, id string) (*entity.User, error)
	CreateFn      func(ctx context.Context, user *entity.User) error
	UpdateFn      func(ctx context.Context, user *entity.User) error
	DeleteFn      func(ctx context.Context, id string) error
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	return m.FindByEmailFn(ctx, email)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	return m.CreateFn(ctx, user)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	return m.UpdateFn(ctx, user)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}
