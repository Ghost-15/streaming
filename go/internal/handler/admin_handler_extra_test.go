package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

var errAdminTest = errors.New("admin test error")

func adminSuspendMock() *mock.MockAdminRepository {
	return &mock.MockAdminRepository{
		SuspendUserFn: func(_ context.Context, _ string, _ bool) error { return nil },
	}
}

func TestAdminHandler_SuspendUser_Query(t *testing.T) {
	h := handler.NewAdminHandler(usecase.NewAdminUseCase(adminSuspendMock()))
	r := newAdminEngine(h)

	for _, q := range []string{"?suspend=true", "?suspend=false"} {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/admin/users/u1/suspend"+q, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("SuspendUser %s status = %d, want 200", q, w.Code)
		}
	}
}

func TestAdminHandler_SuspendUser_JSONBody(t *testing.T) {
	h := handler.NewAdminHandler(usecase.NewAdminUseCase(adminSuspendMock()))
	r := newAdminEngine(h)

	body, _ := json.Marshal(map[string]bool{"suspend": true})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/admin/users/u1/suspend", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("SuspendUser JSON status = %d, want 200", w.Code)
	}
}

func TestAdminHandler_SuspendUser_Error(t *testing.T) {
	repo := &mock.MockAdminRepository{
		SuspendUserFn: func(_ context.Context, _ string, _ bool) error { return errAdminTest },
	}
	h := handler.NewAdminHandler(usecase.NewAdminUseCase(repo))
	r := newAdminEngine(h)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/admin/users/u1/suspend?suspend=true", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("SuspendUser error status = %d, want 500", w.Code)
	}
}

func TestAdminHandler_ErrorPaths(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   []byte
		setup  func(*mock.MockAdminRepository)
	}{
		{
			name: "ListUsers error", method: http.MethodGet, path: "/admin/users",
			setup: func(r *mock.MockAdminRepository) {
				r.ListUsersFn = func(_ context.Context, _, _ int) ([]entity.User, int, error) { return nil, 0, errAdminTest }
			},
		},
		{
			name: "GetUser error", method: http.MethodGet, path: "/admin/users/u1",
			setup: func(r *mock.MockAdminRepository) {
				r.GetUserFn = func(_ context.Context, _ string) (*entity.User, error) { return nil, errAdminTest }
			},
		},
		{
			name: "UpdateUserRole error", method: http.MethodPut, path: "/admin/users/u1/role",
			body: mustJSON(map[string]string{"role": "user"}),
			setup: func(r *mock.MockAdminRepository) {
				r.UpdateUserRoleFn = func(_ context.Context, _ string, _ entity.UserRole) error { return errAdminTest }
			},
		},
		{
			name: "GetStats error", method: http.MethodGet, path: "/admin/stats",
			setup: func(r *mock.MockAdminRepository) {
				r.GetStatsFn = func(_ context.Context) (*entity.AdminStats, error) { return nil, errAdminTest }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockAdminRepository{}
			tt.setup(repo)
			h := handler.NewAdminHandler(usecase.NewAdminUseCase(repo))
			r := newAdminEngine(h)

			var reader *bytes.Reader
			if tt.body != nil {
				reader = bytes.NewReader(tt.body)
			} else {
				reader = bytes.NewReader(nil)
			}
			req := httptest.NewRequestWithContext(context.Background(), tt.method, tt.path, reader)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusInternalServerError {
				t.Errorf("%s status = %d, want 500", tt.name, w.Code)
			}
		})
	}
}

func mustJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
