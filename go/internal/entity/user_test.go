package entity_test

import (
	"testing"

	"github.com/Ghost-15/streaming/internal/entity"
)

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		role entity.UserRole
		want bool
	}{
		{"admin is admin", entity.RoleAdmin, true},
		{"user is not admin", entity.RoleUser, false},
		{"diffuseur is not admin", entity.RoleDiffuseur, false},
		{"anon is not admin", entity.RoleAnon, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &entity.User{Role: tt.role}
			if got := u.IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_IsDiffuseur(t *testing.T) {
	tests := []struct {
		name string
		role entity.UserRole
		want bool
	}{
		{"diffuseur can broadcast", entity.RoleDiffuseur, true},
		{"admin can broadcast", entity.RoleAdmin, true},
		{"user cannot broadcast", entity.RoleUser, false},
		{"anon cannot broadcast", entity.RoleAnon, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &entity.User{Role: tt.role}
			if got := u.IsDiffuseur(); got != tt.want {
				t.Errorf("IsDiffuseur() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_HasRole(t *testing.T) {
	u := &entity.User{Role: entity.RoleUser}

	if !u.HasRole(entity.RoleUser, entity.RoleAdmin) {
		t.Error("expected HasRole to return true for matching role")
	}
	if u.HasRole(entity.RoleAdmin, entity.RoleDiffuseur) {
		t.Error("expected HasRole to return false for non-matching roles")
	}
	if u.HasRole() {
		t.Error("expected HasRole to return false when no roles are provided")
	}
}

func TestUser_FullName(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		want      string
	}{
		{"both set", "Youri", "Emmanuel", "Youri Emmanuel"},
		{"only first", "Youri", "", "Youri "},
		{"only last", "", "Emmanuel", " Emmanuel"},
		{"both empty", "", "", " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &entity.User{FirstName: tt.firstName, LastName: tt.lastName}
			if got := u.FullName(); got != tt.want {
				t.Errorf("FullName() = %q, want %q", got, tt.want)
			}
		})
	}
}
