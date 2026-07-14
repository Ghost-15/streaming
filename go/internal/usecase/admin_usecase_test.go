package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

func TestAdminUseCase_ListUsers(t *testing.T) {
	repo := &mock.MockAdminRepository{
		ListUsersFn: func(_ context.Context, _, _ int) ([]entity.User, int, error) {
			return []entity.User{{ID: "u1"}}, 1, nil
		},
	}
	uc := usecase.NewAdminUseCase(repo)
	users, total, err := uc.ListUsers(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("ListUsers() err = %v", err)
	}
	if total != 1 || len(users) != 1 {
		t.Errorf("ListUsers() total=%d len=%d, want 1/1", total, len(users))
	}
}

func TestAdminUseCase_GetUser(t *testing.T) {
	repo := &mock.MockAdminRepository{
		GetUserFn: func(_ context.Context, id string) (*entity.User, error) {
			return &entity.User{ID: id}, nil
		},
	}
	uc := usecase.NewAdminUseCase(repo)
	u, err := uc.GetUser(context.Background(), "u1")
	if err != nil || u == nil {
		t.Fatalf("GetUser() u=%v err=%v", u, err)
	}
}

func TestAdminUseCase_UpdateUserRole(t *testing.T) {
	tests := []struct {
		name    string
		role    entity.UserRole
		wantErr bool
	}{
		{"valid user", entity.RoleUser, false},
		{"valid diffuseur", entity.RoleDiffuseur, false},
		{"valid admin", entity.RoleAdmin, false},
		{"invalid role", entity.UserRole("superuser"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockAdminRepository{
				UpdateUserRoleFn: func(_ context.Context, _ string, _ entity.UserRole) error { return nil },
			}
			uc := usecase.NewAdminUseCase(repo)
			err := uc.UpdateUserRole(context.Background(), "u1", tt.role)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UpdateUserRole() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAdminUseCase_SuspendUser(t *testing.T) {
	repo := &mock.MockAdminRepository{
		SuspendUserFn: func(_ context.Context, _ string, _ bool) error { return nil },
	}
	uc := usecase.NewAdminUseCase(repo)
	if err := uc.SuspendUser(context.Background(), "u1", true); err != nil {
		t.Fatalf("SuspendUser() err = %v", err)
	}
}

func TestAdminUseCase_GetStats(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mock.MockAdminRepository{
			GetStatsFn: func(_ context.Context) (*entity.AdminStats, error) {
				return &entity.AdminStats{TotalUsers: 5}, nil
			},
		}
		uc := usecase.NewAdminUseCase(repo)
		stats, err := uc.GetStats(context.Background())
		if err != nil || stats.TotalUsers != 5 {
			t.Fatalf("GetStats() stats=%v err=%v", stats, err)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mock.MockAdminRepository{
			GetStatsFn: func(_ context.Context) (*entity.AdminStats, error) { return nil, errors.New("db") },
		}
		uc := usecase.NewAdminUseCase(repo)
		if _, err := uc.GetStats(context.Background()); err == nil {
			t.Error("GetStats() expected error")
		}
	})
}
