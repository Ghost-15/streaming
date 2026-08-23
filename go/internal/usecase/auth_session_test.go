package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

// newSessionUseCase wires an AuthUseCase around a single known user, and returns
// the refresh token store so a test can inspect what was issued and revoked.
func newSessionUseCase(t *testing.T, user *entity.User) (usecase.AuthUseCase, *mock.MockRefreshTokenRepository) {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user.PasswordHash = string(hash)

	deleted := false
	userRepo := &mock.MockUserRepository{
		FindByEmailFn: func(_ context.Context, email string) (*entity.User, error) {
			if deleted || email != user.Email {
				return nil, nil
			}
			return user, nil
		},
		FindByIDFn: func(_ context.Context, id string) (*entity.User, error) {
			if id == user.ID && !deleted {
				return user, nil
			}
			return nil, nil
		},
		DeleteFn: func(_ context.Context, id string) error {
			if id != user.ID {
				return errors.New("user_repo: not found")
			}
			deleted = true
			return nil
		},
	}
	refreshRepo := &mock.MockRefreshTokenRepository{}
	uc := usecase.NewAuthUseCase(userRepo, refreshRepo, testKeyPath, time.Hour, 720*time.Hour)
	return uc, refreshRepo
}

func loginForTest(t *testing.T, uc usecase.AuthUseCase, email string) *usecase.AuthResult {
	t.Helper()

	result, err := uc.Login(context.Background(), email, "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	return result
}

// ─────────────────────────────────────────────────────────────
// Refresh
// ─────────────────────────────────────────────────────────────

func TestAuthUseCase_RefreshRotatesTheToken(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, store := newSessionUseCase(t, user)

	first := loginForTest(t, uc, user.Email)
	second, err := uc.Refresh(context.Background(), first.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if second.RefreshToken == first.RefreshToken {
		t.Error("Refresh() must issue a new refresh token, not reuse the presented one")
	}
	if second.AccessToken == "" {
		t.Error("Refresh() returned an empty access token")
	}
	if second.ExpiresIn != int(time.Hour.Seconds()) {
		t.Errorf("Refresh() expires_in = %d, want %d", second.ExpiresIn, int(time.Hour.Seconds()))
	}
	if len(store.Stored) != 2 {
		t.Fatalf("Refresh() stored %d tokens, want 2", len(store.Stored))
	}
	if store.Stored[0].RevokedAt == nil {
		t.Error("Refresh() must revoke the presented token")
	}
	if store.Stored[1].RevokedAt != nil {
		t.Error("Refresh() must leave the newly issued token usable")
	}
}

func TestAuthUseCase_RefreshRejectsUnknownToken(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, _ := newSessionUseCase(t, user)

	if _, err := uc.Refresh(context.Background(), "not-a-stored-token"); !errors.Is(err, usecase.ErrInvalidRefreshToken) {
		t.Errorf("Refresh() error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestAuthUseCase_RefreshRejectsEmptyToken(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, _ := newSessionUseCase(t, user)

	if _, err := uc.Refresh(context.Background(), ""); !errors.Is(err, usecase.ErrInvalidRefreshToken) {
		t.Errorf("Refresh() error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestAuthUseCase_RefreshRejectsExpiredToken(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, store := newSessionUseCase(t, user)

	result := loginForTest(t, uc, user.Email)
	store.Stored[0].ExpiresAt = time.Now().Add(-time.Minute)

	if _, err := uc.Refresh(context.Background(), result.RefreshToken); !errors.Is(err, usecase.ErrInvalidRefreshToken) {
		t.Errorf("Refresh() error = %v, want ErrInvalidRefreshToken", err)
	}
}

// A rotated token presented a second time means the value leaked: every session
// of that user must be dropped, not just the replayed one.
func TestAuthUseCase_RefreshReplayRevokesEverySession(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, store := newSessionUseCase(t, user)

	first := loginForTest(t, uc, user.Email)
	if _, err := uc.Refresh(context.Background(), first.RefreshToken); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if _, err := uc.Refresh(context.Background(), first.RefreshToken); !errors.Is(err, usecase.ErrInvalidRefreshToken) {
		t.Fatalf("Refresh() replay error = %v, want ErrInvalidRefreshToken", err)
	}
	for i, stored := range store.Stored {
		if stored.RevokedAt == nil {
			t.Errorf("Refresh() replay left token %d usable", i)
		}
	}
}

func TestAuthUseCase_RefreshRejectsSuspendedUser(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, _ := newSessionUseCase(t, user)

	result := loginForTest(t, uc, user.Email)
	suspended := time.Now()
	user.SuspendedAt = &suspended

	if _, err := uc.Refresh(context.Background(), result.RefreshToken); !errors.Is(err, usecase.ErrAccountSuspended) {
		t.Errorf("Refresh() error = %v, want ErrAccountSuspended", err)
	}
}

// ─────────────────────────────────────────────────────────────
// Logout
// ─────────────────────────────────────────────────────────────

func TestAuthUseCase_LogoutRevokesTheToken(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, store := newSessionUseCase(t, user)

	result := loginForTest(t, uc, user.Email)
	if err := uc.Logout(context.Background(), result.RefreshToken); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if store.Stored[0].RevokedAt == nil {
		t.Fatal("Logout() must revoke the refresh token")
	}
	if _, err := uc.Refresh(context.Background(), result.RefreshToken); err == nil {
		t.Error("Refresh() must fail after logout")
	}
}

// Logging out twice must stay safe so a client can always clear its local state.
func TestAuthUseCase_LogoutIsIdempotent(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, _ := newSessionUseCase(t, user)

	result := loginForTest(t, uc, user.Email)
	if err := uc.Logout(context.Background(), result.RefreshToken); err != nil {
		t.Fatalf("Logout() first call error = %v", err)
	}
	if err := uc.Logout(context.Background(), result.RefreshToken); err != nil {
		t.Errorf("Logout() second call error = %v, want nil", err)
	}
}

func TestAuthUseCase_LogoutRejectsEmptyToken(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, _ := newSessionUseCase(t, user)

	if err := uc.Logout(context.Background(), ""); !errors.Is(err, usecase.ErrInvalidRefreshToken) {
		t.Errorf("Logout() error = %v, want ErrInvalidRefreshToken", err)
	}
}

// ─────────────────────────────────────────────────────────────
// Me
// ─────────────────────────────────────────────────────────────

func TestAuthUseCase_MeReturnsTheUserWithoutItsHash(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, _ := newSessionUseCase(t, user)

	got, err := uc.Me(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Me() error = %v", err)
	}
	if got.Email != user.Email {
		t.Errorf("Me() email = %q, want %q", got.Email, user.Email)
	}
	if got.PasswordHash != "" {
		t.Error("Me() must never expose the password hash")
	}
}

func TestAuthUseCase_MeRejectsUnknownUser(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, _ := newSessionUseCase(t, user)

	if _, err := uc.Me(context.Background(), "ghost"); !errors.Is(err, usecase.ErrUserNotFound) {
		t.Errorf("Me() error = %v, want ErrUserNotFound", err)
	}
}

func TestAuthUseCase_MeRejectsEmptyID(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, _ := newSessionUseCase(t, user)

	if _, err := uc.Me(context.Background(), ""); !errors.Is(err, usecase.ErrUserNotFound) {
		t.Errorf("Me() error = %v, want ErrUserNotFound", err)
	}
}

// ─────────────────────────────────────────────────────────────
// DeleteAccount — RGPD right to erasure
// ─────────────────────────────────────────────────────────────

func TestAuthUseCase_DeleteAccountRevokesEverySession(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, store := newSessionUseCase(t, user)

	first := loginForTest(t, uc, user.Email)
	second := loginForTest(t, uc, user.Email)

	if err := uc.DeleteAccount(context.Background(), user.ID); err != nil {
		t.Fatalf("DeleteAccount() error = %v", err)
	}

	for i, stored := range store.Stored {
		if stored.RevokedAt == nil {
			t.Errorf("DeleteAccount() left session %d usable", i)
		}
	}
	if _, err := uc.Refresh(context.Background(), first.RefreshToken); err == nil {
		t.Error("Refresh() must fail once the account is deleted")
	}
	if _, err := uc.Refresh(context.Background(), second.RefreshToken); err == nil {
		t.Error("Refresh() must fail on every session of a deleted account")
	}
}

func TestAuthUseCase_DeleteAccountMakesTheUserUnreadable(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, _ := newSessionUseCase(t, user)

	if err := uc.DeleteAccount(context.Background(), user.ID); err != nil {
		t.Fatalf("DeleteAccount() error = %v", err)
	}
	if _, err := uc.Me(context.Background(), user.ID); !errors.Is(err, usecase.ErrUserNotFound) {
		t.Errorf("Me() error = %v, want ErrUserNotFound", err)
	}
	if _, err := uc.Login(context.Background(), user.Email, "password123"); err == nil {
		t.Error("Login() must fail once the account is deleted")
	}
}

func TestAuthUseCase_DeleteAccountRejectsUnknownUser(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, _ := newSessionUseCase(t, user)

	if err := uc.DeleteAccount(context.Background(), "ghost"); !errors.Is(err, usecase.ErrUserNotFound) {
		t.Errorf("DeleteAccount() error = %v, want ErrUserNotFound", err)
	}
}

func TestAuthUseCase_DeleteAccountRejectsEmptyID(t *testing.T) {
	user := &entity.User{ID: "user-1", Email: "listener@streampulse.fr", Role: entity.RoleUser}
	uc, _ := newSessionUseCase(t, user)

	if err := uc.DeleteAccount(context.Background(), ""); !errors.Is(err, usecase.ErrUserNotFound) {
		t.Errorf("DeleteAccount() error = %v, want ErrUserNotFound", err)
	}
}
