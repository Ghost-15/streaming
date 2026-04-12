package repository

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
)

// UserRepository defines the persistence contract for users.
// Implemented in internal/infrastructure/supabase/user_repo.go
// Never import handler or usecase packages here — dependency rule.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
}
