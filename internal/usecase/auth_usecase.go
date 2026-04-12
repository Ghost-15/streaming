package usecase

import (
	"context"
	"errors"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// AuthUseCase defines the business operations for authentication.
type AuthUseCase interface {
	Register(ctx context.Context, email, password string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (token string, err error)
}

// authUseCase is the concrete implementation injected with its dependencies.
type authUseCase struct {
	userRepo      repository.UserRepository
	jwtPrivateKey string
}

// NewAuthUseCase creates a new AuthUseCase. Called from cmd/server/main.go.
func NewAuthUseCase(userRepo repository.UserRepository, jwtPrivateKey string) AuthUseCase {
	return &authUseCase{
		userRepo:      userRepo,
		jwtPrivateKey: jwtPrivateKey,
	}
}

func (uc *authUseCase) Register(ctx context.Context, email, password string) (*entity.User, error) {
	// TODO Sprint 1 — US-001:
	// 1. Check email uniqueness via userRepo.FindByEmail
	// 2. Hash password with bcrypt
	// 3. Create user with RoleUser
	// 4. Return user
	return nil, errors.New("not implemented")
}

func (uc *authUseCase) Login(ctx context.Context, email, password string) (string, error) {
	// TODO Sprint 1 — US-001:
	// 1. Find user by email
	// 2. Compare bcrypt hash
	// 3. Generate JWT RS256 signed token (exp: 1h)
	// 4. Return token string
	return "", errors.New("not implemented")
}
