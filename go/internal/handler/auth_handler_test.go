package handler_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

// testKeyPath holds the path to the temp RSA private key generated for this test run.
var testKeyPath string

// TestMain generates a throwaway RSA key pair in a temp file so handler tests
// never depend on committed key material or a specific filesystem layout.
func TestMain(m *testing.M) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("test setup: generate RSA key: " + err.Error())
	}

	f, err := os.CreateTemp("", "streampulse-handler-test-private-*.pem")
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
			name:           "empty body",
			body:           map[string]interface{}{},
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
			gin.SetMode(gin.TestMode)
			repo := &mock.MockUserRepository{}

			repo.FindByEmailFn = func(_ context.Context, email string) (*entity.User, error) {
				if tt.body["email"] == "existing@example.com" {
					return &entity.User{ID: "user123", Email: email}, nil
				}
				return nil, nil
			}
			repo.CreateFn = func(_ context.Context, _ *entity.User) error {
				return nil
			}

			uc := usecase.NewAuthUseCase(repo, testKeyPath)
			h := handler.NewAuthHandler(uc)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			engine := gin.New()
			engine.POST("/api/v1/auth/register", h.Register)
			engine.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Register() status = %d, want %d", w.Code, tt.expectedStatus)
			}

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
			name:           "empty body",
			body:           map[string]interface{}{},
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
			gin.SetMode(gin.TestMode)
			repo := &mock.MockUserRepository{}

			repo.FindByEmailFn = func(_ context.Context, email string) (*entity.User, error) {
				if email == "test@example.com" {
					hash, _ := bcrypt.GenerateFromPassword([]byte("ValidPassword123"), bcrypt.DefaultCost)
					return &entity.User{
						ID:           "user123",
						Email:        email,
						PasswordHash: string(hash),
						Role:         entity.RoleUser,
					}, nil
				}
				return nil, nil
			}

			uc := usecase.NewAuthUseCase(repo, testKeyPath)
			h := handler.NewAuthHandler(uc)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			engine := gin.New()
			engine.POST("/api/v1/auth/login", h.Login)
			engine.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Login() status = %d, want %d", w.Code, tt.expectedStatus)
			}

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

// contains is a helper for substring checks in response bodies.
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
