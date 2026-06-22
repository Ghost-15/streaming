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

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

func newAdminEngine(h *handler.AdminHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Inject admin claims into context (simulates RBACMiddleware)
	r.Use(func(c *gin.Context) {
		c.Set("claims", &entity.JWTClaims{
			UserID: "admin-001",
			Email:  "admin@test.com",
			Role:   entity.RoleAdmin,
		})
		c.Next()
	})
	r.GET("/admin/users", h.ListUsers)
	r.GET("/admin/users/:id", h.GetUser)
	r.PUT("/admin/users/:id/role", h.UpdateUserRole)
	r.POST("/admin/users/:id/suspend", h.SuspendUser)
	r.GET("/admin/stats", h.GetStats)
	return r
}

func TestAdminHandler_ListUsers(t *testing.T) {
	repo := &mock.MockAdminRepository{}
	repo.ListUsersFn = func(_ context.Context, page, limit int) ([]entity.User, int, error) {
		return []entity.User{
			{ID: "u1", Email: "a@test.com", Role: entity.RoleUser, CreatedAt: time.Now()},
			{ID: "u2", Email: "b@test.com", Role: entity.RoleDiffuseur, CreatedAt: time.Now()},
		}, 2, nil
	}

	h := handler.NewAdminHandler(usecase.NewAdminUseCase(repo))
	r := newAdminEngine(h)

	req := httptest.NewRequest(http.MethodGet, "/admin/users?page=1&limit=20", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListUsers() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("ListUsers() failed to parse response: %v", err)
	}
	if resp["total"].(float64) != 2 {
		t.Errorf("ListUsers() total = %v, want 2", resp["total"])
	}
}

func TestAdminHandler_GetUser_Found(t *testing.T) {
	repo := &mock.MockAdminRepository{}
	repo.GetUserFn = func(_ context.Context, id string) (*entity.User, error) {
		return &entity.User{ID: id, Email: "user@test.com", Role: entity.RoleUser}, nil
	}

	h := handler.NewAdminHandler(usecase.NewAdminUseCase(repo))
	r := newAdminEngine(h)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/u-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetUser() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAdminHandler_GetUser_NotFound(t *testing.T) {
	repo := &mock.MockAdminRepository{}
	repo.GetUserFn = func(_ context.Context, _ string) (*entity.User, error) {
		return nil, nil
	}

	h := handler.NewAdminHandler(usecase.NewAdminUseCase(repo))
	r := newAdminEngine(h)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/unknown", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetUser() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAdminHandler_UpdateUserRole(t *testing.T) {
	repo := &mock.MockAdminRepository{}
	repo.UpdateUserRoleFn = func(_ context.Context, _ string, _ entity.UserRole) error {
		return nil
	}

	h := handler.NewAdminHandler(usecase.NewAdminUseCase(repo))
	r := newAdminEngine(h)

	body, _ := json.Marshal(map[string]string{"role": "diffuseur"})
	req := httptest.NewRequest(http.MethodPut, "/admin/users/u-123/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UpdateUserRole() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAdminHandler_UpdateUserRole_InvalidRole(t *testing.T) {
	repo := &mock.MockAdminRepository{}
	h := handler.NewAdminHandler(usecase.NewAdminUseCase(repo))
	r := newAdminEngine(h)

	body, _ := json.Marshal(map[string]string{"role": "superuser"})
	req := httptest.NewRequest(http.MethodPut, "/admin/users/u-123/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("UpdateUserRole() invalid role status = %d, want 500", w.Code)
	}
}

func TestAdminHandler_GetStats(t *testing.T) {
	repo := &mock.MockAdminRepository{}
	repo.GetStatsFn = func(_ context.Context) (*entity.AdminStats, error) {
		return &entity.AdminStats{
			TotalUsers: 42,
			ByRole:     map[string]int{"user": 38, "diffuseur": 3, "admin": 1},
		}, nil
	}

	h := handler.NewAdminHandler(usecase.NewAdminUseCase(repo))
	r := newAdminEngine(h)

	req := httptest.NewRequest(http.MethodGet, "/admin/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetStats() status = %d, want %d", w.Code, http.StatusOK)
	}

	var stats entity.AdminStats
	if err := json.Unmarshal(w.Body.Bytes(), &stats); err != nil {
		t.Fatalf("GetStats() failed to parse response: %v", err)
	}
	if stats.TotalUsers != 42 {
		t.Errorf("GetStats() total_users = %d, want 42", stats.TotalUsers)
	}
}
