package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

// TestAuthHandler_Register tests the POST /auth/register endpoint.
// Sprint 1 — US-001 — STR-13.
func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		expectedError  string // substring check in response error
	}{
		{
			name: "valid registration",
			body: map[string]interface{}{
				"email":    "user@example.com",
				"password": "ValidPassword123",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid email",
			body: map[string]interface{}{
				"email":    "not-an-email",
				"password": "ValidPassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email",
		},
		{
			name: "password too short",
			body: map[string]interface{}{
				"email":    "user@example.com",
				"password": "short",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "min",
		},
		{
			name: "missing email",
			body: map[string]interface{}{
				"password": "ValidPassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "required",
		},
		{
			name: "missing password",
			body: map[string]interface{}{
				"email": "user@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "required",
		},
		{
			name: "empty body",
			body: map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "required",
		},
		{
			name: "email already exists",
			body: map[string]interface{}{
				"email":    "existing@example.com",
				"password": "ValidPassword123",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "already",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			repo := &mock.MockUserRepository{}
			uc := usecase.NewAuthUseCase(repo, "./testdata/private.pem")
			h := handler.NewAuthHandler(uc)

			// Marshal body
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			engine := gin.New()
			engine.POST("/api/v1/auth/register", h.Register)

			// Execute
			engine.ServeHTTP(w, req)

			// Assert status
			if w.Code != tt.expectedStatus {
				t.Errorf("Register() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			// Assert error message in response if expected
			if tt.expectedError != "" {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("Register() failed to parse response: %v", err)
					return
				}

				if errMsg, ok := resp["error"]; ok {
					if !contains(errMsg.(string), tt.expectedError) {
						t.Errorf("Register() error = %q, expected to contain %q", errMsg, tt.expectedError)
					}
				} else {
					t.Errorf("Register() expected error field in response, got none")
				}
			}
		})
	}
}

// TestAuthHandler_Login tests the POST /auth/login endpoint.
// Sprint 1 — US-001 — STR-13.
func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		expectedError  string
		checkToken     bool
	}{
		{
			name: "valid login",
			body: map[string]interface{}{
				"email":    "test@example.com",
				"password": "ValidPassword123",
			},
			expectedStatus: http.StatusOK,
			checkToken:     true,
		},
		{
			name: "invalid email format",
			body: map[string]interface{}{
				"email":    "not-an-email",
				"password": "ValidPassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email",
		},
		{
			name: "password too short",
			body: map[string]interface{}{
				"email":    "test@example.com",
				"password": "short",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "min",
		},
		{
			name: "missing email",
			body: map[string]interface{}{
				"password": "ValidPassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "required",
		},
		{
			name: "missing password",
			body: map[string]interface{}{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "required",
		},
		{
			name: "empty body",
			body: map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "required",
		},
		{
			name: "invalid credentials",
			body: map[string]interface{}{
				"email":    "wrong@example.com",
				"password": "WrongPassword123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			repo := &mock.MockUserRepository{}
			uc := usecase.NewAuthUseCase(repo, "./testdata/private.pem")
			h := handler.NewAuthHandler(uc)

			// Marshal body
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			engine := gin.New()
			engine.POST("/api/v1/auth/login", h.Login)

			// Execute
			engine.ServeHTTP(w, req)

			// Assert status
			if w.Code != tt.expectedStatus {
				t.Errorf("Login() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			// Assert error message in response if expected
			if tt.expectedError != "" {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("Login() failed to parse response: %v", err)
					return
				}

				if errMsg, ok := resp["error"]; ok {
					if !contains(errMsg.(string), tt.expectedError) {
						t.Errorf("Login() error = %q, expected to contain %q", errMsg, tt.expectedError)
					}
				} else {
					t.Errorf("Login() expected error field in response, got none")
				}
			}

			// Assert token presence if expected
			if tt.checkToken && w.Code == http.StatusOK {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("Login() failed to parse response: %v", err)
					return
				}

				if _, ok := resp["token"]; !ok {
					t.Errorf("Login() expected token in response, got none")
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
