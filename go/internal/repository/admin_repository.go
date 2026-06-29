package repository

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
)

// AdminRepository defines the persistence contract for admin operations.
type AdminRepository interface {
	ListUsers(ctx context.Context, page, limit int) ([]entity.User, int, error)
	GetUser(ctx context.Context, id string) (*entity.User, error)
	UpdateUserRole(ctx context.Context, id string, role entity.UserRole) error
	SuspendUser(ctx context.Context, id string, suspend bool) error
	GetStats(ctx context.Context) (*entity.AdminStats, error)
}
