package usecase_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

// testKeyPath holds the path to the temp RSA private key generated for this test run.
var testKeyPath string

// TestMain generates a throwaway RSA key pair in a temp file so tests never
// depend on committed key material. The file is deleted after the suite runs.
func TestMain(m *testing.M) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("test setup: generate RSA key: " + err.Error())
	}

	f, err := os.CreateTemp("", "streampulse-test-private-*.pem")
	if err != nil {
		panic("test setup: create temp key file: " + err.Error())
	}

	if err := pem.Encode(f, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		panic("test setup: encode RSA key: " + err.Error())
	}
	f.Close()

	testKeyPath = f.Name()
	code := m.Run()
	os.Remove(testKeyPath)
	os.Exit(code)
}

// ─────────────────────────────────────────────────────────────
// Register
// ─────────────────────────────────────────────────────────────

func TestAuthUseCase_Register(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		password   string
		repoSetup  func(*mock.MockUserRepository)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "success",
			email:    "new@example.com",
			password: "password123",
			repoSetup: func(r *mock.MockUserRepository) {
				r.FindByEmailFn = func(_ context.Context, _ string) (*entity.User, error) {
					return nil, nil
				}
				r.CreateFn = func(_ context.Context, u *entity.User) error {
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
			uc := usecase.NewAuthUseCase(repo, testKeyPath)

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
					return nil, nil
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
			uc := usecase.NewAuthUseCase(repo, testKeyPath)

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
				if len(strings.Split(token, ".")) != 3 {
					t.Errorf("Login() token %q does not look like a JWT", token)
				}
			}
		})
	}
}
