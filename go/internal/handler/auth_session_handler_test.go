package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

// newSessionEngine wires the session routes around a single known user. When
// authenticated is false no claims are injected, which is how the RBAC
// middleware leaves the context for an anonymous caller.
func newSessionEngine(t *testing.T, authenticated bool) (*gin.Engine, *entity.User) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := &entity.User{
		ID:           "user-1",
		Email:        "listener@streampulse.fr",
		Role:         entity.RoleUser,
		PasswordHash: string(hash),
	}

	deleted := false
	userRepo := &mock.MockUserRepository{
		FindByEmailFn: func(_ context.Context, email string) (*entity.User, error) {
			if deleted || email != user.Email {
				return nil, nil
			}
			return user, nil
		},
		FindByIDFn: func(_ context.Context, id string) (*entity.User, error) {
			if deleted || id != user.ID {
				return nil, nil
			}
			return user, nil
		},
		UpdatePasswordFn: func(_ context.Context, _, passwordHash string) error {
			user.PasswordHash = passwordHash
			return nil
		},
		DeleteFn: func(_ context.Context, _ string) error {
			deleted = true
			return nil
		},
	}

	uc := usecase.NewAuthUseCase(userRepo, &mock.MockRefreshTokenRepository{}, testKeyPath, time.Hour, 720*time.Hour)
	h := handler.NewAuthHandler(uc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		if authenticated {
			c.Set("claims", &entity.JWTClaims{
				UserID: user.ID,
				Email:  user.Email,
				Role:   user.Role,
			})
		}
		c.Next()
	})
	r.POST("/auth/login", h.Login)
	r.POST("/auth/refresh", h.Refresh)
	r.POST("/auth/logout", h.Logout)
	r.GET("/auth/me", h.Me)
	r.PUT("/auth/password", h.ChangePassword)
	r.DELETE("/auth/me", h.DeleteMe)
	return r, user
}

func doSessionJSON(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequestWithContext(context.Background(), method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func loginTokens(t *testing.T, r *gin.Engine) map[string]any {
	t.Helper()

	w := doSessionJSON(t, r, http.MethodPost, "/auth/login", map[string]string{
		"email":    "listener@streampulse.fr",
		"password": "password123",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200 (body %s)", w.Code, w.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	return payload
}

func TestAuthHandler_LoginReturnsARefreshToken(t *testing.T) {
	r, _ := newSessionEngine(t, false)
	payload := loginTokens(t, r)

	for _, key := range []string{"token", "refresh_token", "expires_in", "user"} {
		if _, ok := payload[key]; !ok {
			t.Errorf("login response is missing %q", key)
		}
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	r, _ := newSessionEngine(t, false)
	payload := loginTokens(t, r)

	w := doSessionJSON(t, r, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": payload["refresh_token"].(string),
	})
	if w.Code != http.StatusOK {
		t.Fatalf("Refresh() status = %d, want 200 (body %s)", w.Code, w.Body.String())
	}
}

func TestAuthHandler_RefreshRejectsAnUnknownToken(t *testing.T) {
	r, _ := newSessionEngine(t, false)

	w := doSessionJSON(t, r, http.MethodPost, "/auth/refresh", map[string]string{"refresh_token": "nope"})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Refresh() status = %d, want 401", w.Code)
	}
}

func TestAuthHandler_RefreshRejectsAMalformedBody(t *testing.T) {
	r, _ := newSessionEngine(t, false)

	w := doSessionJSON(t, r, http.MethodPost, "/auth/refresh", map[string]string{})
	if w.Code != http.StatusBadRequest {
		t.Errorf("Refresh() status = %d, want 400", w.Code)
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	r, _ := newSessionEngine(t, false)
	payload := loginTokens(t, r)

	w := doSessionJSON(t, r, http.MethodPost, "/auth/logout", map[string]string{
		"refresh_token": payload["refresh_token"].(string),
	})
	if w.Code != http.StatusNoContent {
		t.Errorf("Logout() status = %d, want 204", w.Code)
	}
}

func TestAuthHandler_LogoutRejectsAMalformedBody(t *testing.T) {
	r, _ := newSessionEngine(t, false)

	w := doSessionJSON(t, r, http.MethodPost, "/auth/logout", map[string]string{})
	if w.Code != http.StatusBadRequest {
		t.Errorf("Logout() status = %d, want 400", w.Code)
	}
}

func TestAuthHandler_Me(t *testing.T) {
	r, user := newSessionEngine(t, true)

	w := doSessionJSON(t, r, http.MethodGet, "/auth/me", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("Me() status = %d, want 200 (body %s)", w.Code, w.Body.String())
	}

	var payload map[string]map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode me response: %v", err)
	}
	if payload["user"]["email"] != user.Email {
		t.Errorf("Me() email = %v, want %q", payload["user"]["email"], user.Email)
	}
	if _, leaked := payload["user"]["password_hash"]; leaked {
		t.Error("Me() must never serialise the password hash")
	}
}

func TestAuthHandler_MeRequiresClaims(t *testing.T) {
	r, _ := newSessionEngine(t, false)

	w := doSessionJSON(t, r, http.MethodGet, "/auth/me", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Me() status = %d, want 401", w.Code)
	}
}

func TestAuthHandler_ChangePassword(t *testing.T) {
	r, _ := newSessionEngine(t, true)

	w := doSessionJSON(t, r, http.MethodPut, "/auth/password", map[string]string{
		"current_password": "password123",
		"new_password":     "new-password-456",
	})
	if w.Code != http.StatusNoContent {
		t.Fatalf("ChangePassword() status = %d, want 204 (body %s)", w.Code, w.Body.String())
	}
}

func TestAuthHandler_ChangePasswordRejectsAWrongCurrentPassword(t *testing.T) {
	r, _ := newSessionEngine(t, true)

	w := doSessionJSON(t, r, http.MethodPut, "/auth/password", map[string]string{
		"current_password": "not-the-password",
		"new_password":     "new-password-456",
	})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("ChangePassword() status = %d, want 401", w.Code)
	}
}

func TestAuthHandler_ChangePasswordRejectsAnIdenticalPassword(t *testing.T) {
	r, _ := newSessionEngine(t, true)

	w := doSessionJSON(t, r, http.MethodPut, "/auth/password", map[string]string{
		"current_password": "password123",
		"new_password":     "password123",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("ChangePassword() status = %d, want 400", w.Code)
	}
}

func TestAuthHandler_ChangePasswordRequiresClaims(t *testing.T) {
	r, _ := newSessionEngine(t, false)

	w := doSessionJSON(t, r, http.MethodPut, "/auth/password", map[string]string{
		"current_password": "password123",
		"new_password":     "new-password-456",
	})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("ChangePassword() status = %d, want 401", w.Code)
	}
}

func TestAuthHandler_ChangePasswordRejectsAShortPassword(t *testing.T) {
	r, _ := newSessionEngine(t, true)

	w := doSessionJSON(t, r, http.MethodPut, "/auth/password", map[string]string{
		"current_password": "password123",
		"new_password":     "short",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("ChangePassword() status = %d, want 400", w.Code)
	}
}

func TestAuthHandler_DeleteMe(t *testing.T) {
	r, _ := newSessionEngine(t, true)

	w := doSessionJSON(t, r, http.MethodDelete, "/auth/me", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("DeleteMe() status = %d, want 204 (body %s)", w.Code, w.Body.String())
	}

	// The account is gone, so reading the profile must now 404.
	w = doSessionJSON(t, r, http.MethodGet, "/auth/me", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("Me() after deletion status = %d, want 404", w.Code)
	}
}

func TestAuthHandler_DeleteMeRequiresClaims(t *testing.T) {
	r, _ := newSessionEngine(t, false)

	w := doSessionJSON(t, r, http.MethodDelete, "/auth/me", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("DeleteMe() status = %d, want 401", w.Code)
	}
}
