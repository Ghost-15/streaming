package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// ErrEmailAlreadyInUse is returned when registration email is already registered.
var ErrEmailAlreadyInUse = errors.New("auth: email already in use")

// AuthUseCase defines the business operations for authentication.
type AuthUseCase interface {
	Register(ctx context.Context, email, password string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (token string, err error)
}

// authUseCase is the concrete implementation injected with its dependencies.
type authUseCase struct {
	userRepo      repository.UserRepository
	jwtPrivateKey string // path to the RSA private key PEM file
}

// NewAuthUseCase creates a new AuthUseCase. Called from cmd/server/main.go.
func NewAuthUseCase(userRepo repository.UserRepository, jwtPrivateKey string) AuthUseCase {
	return &authUseCase{
		userRepo:      userRepo,
		jwtPrivateKey: jwtPrivateKey,
	}
}

// Register creates a new user account with role "user".
// Returns an error if the email is already in use.
func (uc *authUseCase) Register(ctx context.Context, email, password string) (*entity.User, error) {
	// 1. Check email uniqueness.
	existing, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("auth: register: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyInUse
	}

	// 2. Hash password with bcrypt (cost 12 — secure enough, fast enough for tests).
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("auth: register: hash password: %w", err)
	}

	// 3. Persist the new user; DB generates UUID and created_at.
	user := &entity.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         entity.RoleUser,
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("auth: register: %w", err)
	}
	return user, nil
}

// Login verifies credentials and returns a signed JWT RS256 token (TTL 1 h).
// Always returns the same generic error for wrong email or wrong password to avoid
// user enumeration (RFC 6749 §5.2 best practice).
func (uc *authUseCase) Login(ctx context.Context, email, password string) (string, error) {
	const errInvalid = "auth: invalid credentials"

	// 1. Look up user.
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("auth: login: %w", err)
	}
	if user == nil {
		return "", errors.New(errInvalid)
	}

	// 2. Constant-time password check.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New(errInvalid)
	}

	// 3. Load the RSA private key from disk.
	keyBytes, err := os.ReadFile(uc.jwtPrivateKey)
	if err != nil {
		return "", fmt.Errorf("auth: login: read private key: %w", err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		return "", fmt.Errorf("auth: login: parse private key: %w", err)
	}

	// 4. Build and sign the JWT (RS256, exp 1 h).
	claims := entity.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("auth: login: sign token: %w", err)
	}
	return token, nil
}
