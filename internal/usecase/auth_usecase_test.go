package usecase_test

import (
	"context"
	"testing"

	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

// TestAuthUseCase_Login tests the Login use case.
// Table-driven pattern — ajouter des cas au fur et à mesure (Sprint 1 — US-001).
func TestAuthUseCase_Login(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		wantErr   bool
	}{
		{
			name:     "not implemented yet",
			email:    "test@example.com",
			password: "password123",
			wantErr:  true, // stub retourne toujours une erreur
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockUserRepository{}
			uc := usecase.NewAuthUseCase(repo, "./testdata/private.pem")

			_, err := uc.Login(context.Background(), tt.email, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAuthUseCase_Register tests the Register use case.
func TestAuthUseCase_Register(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{
			name:     "not implemented yet",
			email:    "new@example.com",
			password: "password123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockUserRepository{}
			uc := usecase.NewAuthUseCase(repo, "./testdata/private.pem")

			_, err := uc.Register(context.Background(), tt.email, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
