package usecase

import (
	"context"
	"fmt"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// AdminUseCase defines the business operations for admin management.
type AdminUseCase interface {
	ListUsers(ctx context.Context, page, limit int) ([]entity.User, int, error)
	GetUser(ctx context.Context, id string) (*entity.User, error)
	UpdateUserRole(ctx context.Context, id string, role entity.UserRole) error
	SuspendUser(ctx context.Context, id string, suspend bool) error
	GetStats(ctx context.Context) (*entity.AdminStats, error)
}

type adminUseCase struct {
	adminRepo repository.AdminRepository
}

// NewAdminUseCase creates a new AdminUseCase.
func NewAdminUseCase(adminRepo repository.AdminRepository) AdminUseCase {
	return &adminUseCase{adminRepo: adminRepo}
}

func (uc *adminUseCase) ListUsers(ctx context.Context, page, limit int) ([]entity.User, int, error) {
	users, total, err := uc.adminRepo.ListUsers(ctx, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("admin: list users: %w", err)
	}
	return users, total, nil
}

func (uc *adminUseCase) GetUser(ctx context.Context, id string) (*entity.User, error) {
	user, err := uc.adminRepo.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("admin: get user: %w", err)
	}
	return user, nil
}

func (uc *adminUseCase) UpdateUserRole(ctx context.Context, id string, role entity.UserRole) error {
	if role != entity.RoleUser && role != entity.RoleDiffuseur && role != entity.RoleAdmin {
		return fmt.Errorf("admin: invalid role %q", role)
	}
	return uc.adminRepo.UpdateUserRole(ctx, id, role)
}

func (uc *adminUseCase) SuspendUser(ctx context.Context, id string, suspend bool) error {
	return uc.adminRepo.SuspendUser(ctx, id, suspend)
}

func (uc *adminUseCase) GetStats(ctx context.Context) (*entity.AdminStats, error) {
	stats, err := uc.adminRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("admin: get stats: %w", err)
	}
	return stats, nil
}
