package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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

var ErrAccountSuspended = errors.New("auth: account suspended")

// ErrInvalidRefreshToken is returned when a refresh token is unknown, expired or
// already revoked. The same error covers all three cases so a caller cannot probe
// which refresh tokens exist.
var ErrInvalidRefreshToken = errors.New("auth: invalid refresh token")

// ErrUserNotFound is returned when the subject of a valid token no longer exists.
var ErrUserNotFound = errors.New("auth: user not found")

// ErrInvalidCredentials is returned when a password check fails outside of login.
var ErrInvalidCredentials = errors.New("auth: invalid credentials")

// ErrPasswordTooShort mirrors the minimum enforced by the registration binding.
var ErrPasswordTooShort = errors.New("auth: password too short")

// ErrPasswordUnchanged is returned when the new password equals the current one.
var ErrPasswordUnchanged = errors.New("auth: password unchanged")

// minPasswordLength matches the `min=8` binding used on register and login.
const minPasswordLength = 8

// AuthResult carries the credentials issued by a successful authentication.
// ExpiresIn is the access token lifetime in seconds, so a client can schedule a
// refresh instead of waiting for a 401.
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	User         *entity.User
}

// AuthUseCase defines the business operations for authentication.
type AuthUseCase interface {
	Register(ctx context.Context, email, password, firstName, lastName, role string) (*AuthResult, error)
	Login(ctx context.Context, email, password string) (*AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (*AuthResult, error)
	Logout(ctx context.Context, refreshToken string) error
	Me(ctx context.Context, userID string) (*entity.User, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	DeleteAccount(ctx context.Context, userID string) error
}

// authUseCase is the concrete implementation injected with its dependencies.
type authUseCase struct {
	userRepo      repository.UserRepository
	refreshRepo   repository.RefreshTokenRepository
	jwtPrivateKey string // path to the RSA private key PEM file
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// NewAuthUseCase creates a new AuthUseCase. Called from cmd/server/main.go.
func NewAuthUseCase(
	userRepo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	jwtPrivateKey string,
	accessTTL, refreshTTL time.Duration,
) AuthUseCase {
	return &authUseCase{
		userRepo:      userRepo,
		refreshRepo:   refreshRepo,
		jwtPrivateKey: jwtPrivateKey,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

// Register creates a new user account with the given role (user or diffuseur).
// Returns an error if the email is already in use.
func (uc *authUseCase) Register(ctx context.Context, email, password, firstName, lastName, role string) (*AuthResult, error) {
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
	userRole := entity.RoleUser
	if role == string(entity.RoleDiffuseur) {
		userRole = entity.RoleDiffuseur
	}
	user := &entity.User{
		Email:        email,
		PasswordHash: string(hash),
		FirstName:    firstName,
		LastName:     lastName,
		Role:         userRole,
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("auth: register: %w", err)
	}
	return uc.issueSession(ctx, user)
}

// Login verifies credentials and returns a signed JWT RS256 access token plus a
// rotating refresh token.
// Always returns the same generic error for wrong email or wrong password to avoid
// user enumeration (RFC 6749 5.2 best practice).
func (uc *authUseCase) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	const errInvalid = "auth: invalid credentials"

	// 1. Look up user.
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("auth: login: %w", err)
	}
	if user == nil {
		return nil, errors.New(errInvalid)
	}

	// 2. Constant-time password check.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New(errInvalid)
	}

	if user.IsSuspended() {
		return nil, ErrAccountSuspended
	}

	return uc.issueSession(ctx, user)
}

// Refresh exchanges a valid refresh token for a new pair, revoking the presented
// one. Presenting an already revoked token is treated as a replay: every session
// of that user is revoked, because the value has likely leaked.
func (uc *authUseCase) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	if refreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	stored, err := uc.refreshRepo.FindByHash(ctx, hashRefreshToken(refreshToken))
	if err != nil {
		return nil, fmt.Errorf("auth: refresh: %w", err)
	}
	if stored == nil {
		return nil, ErrInvalidRefreshToken
	}

	if stored.IsRevoked() {
		// Replay of a rotated token — assume the value leaked and drop every session.
		if revokeErr := uc.refreshRepo.RevokeAllForUser(ctx, stored.UserID); revokeErr != nil {
			return nil, fmt.Errorf("auth: refresh: %w", revokeErr)
		}
		return nil, ErrInvalidRefreshToken
	}
	if stored.IsExpired(time.Now()) {
		return nil, ErrInvalidRefreshToken
	}

	user, err := uc.userRepo.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, fmt.Errorf("auth: refresh: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidRefreshToken
	}
	if user.IsSuspended() {
		return nil, ErrAccountSuspended
	}

	// Rotation: the presented token is single use.
	if err := uc.refreshRepo.Revoke(ctx, stored.TokenHash); err != nil {
		return nil, fmt.Errorf("auth: refresh: %w", err)
	}
	return uc.issueSession(ctx, user)
}

// Logout revokes a refresh token. Revoking an unknown or already revoked token
// succeeds, so a client can always clear its local state.
func (uc *authUseCase) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return ErrInvalidRefreshToken
	}
	if err := uc.refreshRepo.Revoke(ctx, hashRefreshToken(refreshToken)); err != nil {
		return fmt.Errorf("auth: logout: %w", err)
	}
	return nil
}

// Me returns the current state of the authenticated user. It reads the database
// rather than trusting the token claims, so a role change or a suspension applies
// before the access token expires.
func (uc *authUseCase) Me(ctx context.Context, userID string) (*entity.User, error) {
	if userID == "" {
		return nil, ErrUserNotFound
	}
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("auth: me: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return withoutPasswordHash(user), nil
}

// ChangePassword replaces a user credential after re-checking the current one.
// Every session is revoked afterwards: a password change is how a user reacts to
// a suspected compromise, so tokens issued before it must stop working.
func (uc *authUseCase) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if userID == "" {
		return ErrUserNotFound
	}
	if len(newPassword) < minPasswordLength {
		return ErrPasswordTooShort
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("auth: change password: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}
	if currentPassword == newPassword {
		return ErrPasswordUnchanged
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("auth: change password: hash password: %w", err)
	}
	if err := uc.userRepo.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return fmt.Errorf("auth: change password: %w", err)
	}
	if err := uc.refreshRepo.RevokeAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("auth: change password: %w", err)
	}
	return nil
}

// DeleteAccount erases a user on their own request (RGPD right to erasure).
// Sessions are revoked first so a valid refresh token cannot outlive the account.
// Streams, playlists and favorites cascade with the row; listen history is kept
// but detached, so aggregate statistics survive without staying attributable.
func (uc *authUseCase) DeleteAccount(ctx context.Context, userID string) error {
	if userID == "" {
		return ErrUserNotFound
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("auth: delete account: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := uc.refreshRepo.RevokeAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("auth: delete account: %w", err)
	}
	if err := uc.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("auth: delete account: %w", err)
	}
	return nil
}

// issueSession mints an access token and a refresh token for an authenticated user.
func (uc *authUseCase) issueSession(ctx context.Context, user *entity.User) (*AuthResult, error) {
	accessToken, err := uc.signToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := newRefreshToken()
	if err != nil {
		return nil, err
	}
	stored := &entity.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashRefreshToken(refreshToken),
		ExpiresAt: time.Now().Add(uc.refreshTTL),
	}
	if err := uc.refreshRepo.Create(ctx, stored); err != nil {
		return nil, fmt.Errorf("auth: issue session: %w", err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(uc.accessTTL.Seconds()),
		User:         withoutPasswordHash(user),
	}, nil
}

// withoutPasswordHash copies a user with its credential stripped. Returning a
// copy keeps the entity owned by the repository untouched.
func withoutPasswordHash(user *entity.User) *entity.User {
	safe := *user
	safe.PasswordHash = ""
	return &safe
}

func (uc *authUseCase) signToken(user *entity.User) (string, error) {
	// Load the RSA private key from disk.
	keyBytes, err := os.ReadFile(uc.jwtPrivateKey)
	if err != nil {
		return "", fmt.Errorf("auth: sign token: read private key: %w", err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		return "", fmt.Errorf("auth: sign token: parse private key: %w", err)
	}

	// Build and sign the JWT (RS256, lifetime from AUTH_ACCESS_TOKEN_TTL).
	claims := entity.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(uc.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("auth: sign token: %w", err)
	}
	return token, nil
}

// newRefreshToken returns 256 bits of cryptographic randomness, URL-safe encoded.
func newRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("auth: new refresh token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// hashRefreshToken derives the value stored in the database. The token is high
// entropy already, so a plain SHA-256 is enough — no salt or KDF needed.
func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
