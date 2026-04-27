package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

const (
	testPrivateKey = "./testdata/private.pem"
	testPublicKey  = "./testdata/public.pem" //nolint:unused // referenced in rbac tests
)

// ─────────────────────────────────────────────────────────────
// Register
// ─────────────────────────────────────────────────────────────

func TestAuthUseCase_Register(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		repoSetup   func(*mock.MockUserRepository)
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name:     "success",
			email:    "new@example.com",
			password: "password123",
			repoSetup: func(r *mock.MockUserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return nil, nil // no existing user
				}
				r.CreateFn = func(_ context.Context, u *entity.User) error {
					// Simulate DB writing back the generated UUID.
					u.ID = "generated-uuid"
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:     "email already in use",
			email:    "existing@example.com",
			password: "password123",
			repoSetup: func(r *mock.MockUserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return &entity.User{Email: "existing@example.com"}, nil
				}
			},
			wantErr:    true,
			wantErrMsg: "email already in use",
		},
		{
			name:     "repository error on find",
			email:    "fail@example.com",
			password: "password123",
			repoSetup: func(r *mock.MockUserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return nil, errors.New("db down")
				}
			},
			wantErr: true,
		},
		{
			name:     "repository error on create",
			email:    "new@example.com",
			password: "password123",
			repoSetup: func(r *mock.MockUserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return nil, nil
				}
				r.CreateFn = func(_ context.Context, _ *entity.User) error {
					return errors.New("insert failed")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockUserRepository{}
			tt.repoSetup(repo)
			uc := usecase.NewAuthUseCase(repo, testPrivateKey)

			user, err := uc.Register(context.Background(), tt.email, tt.password)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("Register() error = %q, want to contain %q", err.Error(), tt.wantErrMsg)
			}
			if !tt.wantErr {
				if user == nil {
					t.Fatal("Register() returned nil user on success")
				}
				if user.Role != entity.RoleUser {
					t.Errorf("Register() role = %q, want %q", user.Role, entity.RoleUser)
				}
				if user.PasswordHash == "" {
					t.Error("Register() password hash must not be empty")
				}
				if user.PasswordHash == "password123" {
					t.Error("Register() password must be hashed, not stored in plain text")
				}
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// Login
// ─────────────────────────────────────────────────────────────

func TestAuthUseCase_Login(t *testing.T) {
	// Pre-hash a known password for the success case.
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("test setup: bcrypt: %v", err)
	}

	validUser := &entity.User{
		ID:           "user-uuid-123",
		Email:        "user@example.com",
		PasswordHash: string(hash),
		Role:         entity.RoleUser,
	}

	tests := []struct {
		name       string
		email      string
		password   string
		repoSetup  func(*mock.MockUserRepository)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "success — returns non-empty JWT",
			email:    "user@example.com",
			password: "correct-password",
			repoSetup: func(r *mock.MockUserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return validUser, nil
				}
			},
			wantErr: false,
		},
		{
			name:     "user not found",
			email:    "ghost@example.com",
			password: "anything",
			repoSetup: func(r *mock.MockUserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return nil, nil // user does not exist
				}
			},
			wantErr:    true,
			wantErrMsg: "invalid credentials",
		},
		{
			name:     "wrong password",
			email:    "user@example.com",
			password: "wrong-password",
			repoSetup: func(r *mock.MockUserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return validUser, nil
				}
			},
			wantErr:    true,
			wantErrMsg: "invalid credentials",
		},
		{
			name:     "repository error",
			email:    "user@example.com",
			password: "correct-password",
			repoSetup: func(r *mock.MockUserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return nil, errors.New("db timeout")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockUserRepository{}
			tt.repoSetup(repo)
			uc := usecase.NewAuthUseCase(repo, testPrivateKey)

			token, err := uc.Login(context.Background(), tt.email, tt.password)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("Login() error = %q, want to contain %q", err.Error(), tt.wantErrMsg)
			}
			if !tt.wantErr {
				if token == "" {
					t.Error("Login() returned empty token on success")
				}
				// A JWT has exactly 3 base64url parts separated by dots.
				if len(strings.Split(token, ".")) != 3 {
					t.Errorf("Login() token %q does not look like a JWT", token)
				}
			}
		})
	}
}
