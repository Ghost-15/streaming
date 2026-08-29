package router_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Ghost-15/streaming/internal/config"
	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/router"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

// This file is the security half of the iterative test plan: it exercises the
// middleware chain the way an attacker would, rather than the way a client does.

// signToken mints an access token accepted by the router built with key.
func signToken(t *testing.T, key *rsa.PrivateKey, role entity.UserRole, expiry time.Time) string {
	t.Helper()

	claims := entity.JWTClaims{
		UserID: "user-1",
		Email:  "listener@streampulse.fr",
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func call(t *testing.T, engine *gin.Engine, method, path, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

// protectedRoutes lists one representative route per privilege level.
var protectedRoutes = []struct {
	method string
	path   string
	// minimum role able to reach the handler
	requires entity.UserRole
}{
	{method: http.MethodGet, path: "/api/v1/playlists", requires: entity.RoleUser},
	{method: http.MethodGet, path: "/api/v1/favorites", requires: entity.RoleUser},
	{method: http.MethodGet, path: "/api/v1/recommendations", requires: entity.RoleUser},
	{method: http.MethodGet, path: "/api/v1/auth/me", requires: entity.RoleUser},
	{method: http.MethodGet, path: "/api/v1/streams/mine", requires: entity.RoleDiffuseur},
	{method: http.MethodPost, path: "/api/v1/streams", requires: entity.RoleDiffuseur},
	{method: http.MethodGet, path: "/api/v1/admin/users", requires: entity.RoleAdmin},
	{method: http.MethodGet, path: "/api/v1/admin/stats", requires: entity.RoleAdmin},
}

// ─────────────────────────────────────────────────────────────
// Authentication
// ─────────────────────────────────────────────────────────────

func TestSecurity_ProtectedRoutesRejectAnonymousCallers(t *testing.T) {
	engine, _ := buildTestRouter(t, true)

	for _, route := range protectedRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			w := call(t, engine, route.method, route.path, "")
			if w.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401 without a token", w.Code)
			}
		})
	}
}

func TestSecurity_ProtectedRoutesRejectMalformedHeaders(t *testing.T) {
	engine, key := buildTestRouter(t, true)
	valid := signToken(t, key, entity.RoleAdmin, time.Now().Add(time.Hour))

	headers := map[string]string{
		"empty":                "",
		"scheme without token": "Bearer ",
		"wrong scheme":         "Basic " + valid,
		"no scheme":            valid,
		"lowercase scheme":     "bearer " + valid,
		"garbage":              "Bearer not-a-jwt",
		"two dots only":        "Bearer ..",
	}

	for name, header := range headers {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/admin/users", nil)
			if header != "" {
				req.Header.Set("Authorization", header)
			}
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401 for header %q", w.Code, header)
			}
		})
	}
}

func TestSecurity_ExpiredTokenIsRejected(t *testing.T) {
	engine, key := buildTestRouter(t, true)
	expired := signToken(t, key, entity.RoleAdmin, time.Now().Add(-time.Minute))

	w := call(t, engine, http.MethodGet, "/api/v1/admin/users", expired)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 for an expired token", w.Code)
	}
}

// A token signed by a different key must not be accepted, otherwise anybody
// able to run the code could mint admin sessions.
func TestSecurity_TokenSignedByAnotherKeyIsRejected(t *testing.T) {
	engine, _ := buildTestRouter(t, true)

	foreignKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate foreign key: %v", err)
	}
	forged := signToken(t, foreignKey, entity.RoleAdmin, time.Now().Add(time.Hour))

	w := call(t, engine, http.MethodGet, "/api/v1/admin/users", forged)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 for a token signed by an unknown key", w.Code)
	}
}

// Algorithm confusion: the attacker re-signs the payload with HS256, using the
// public key bytes as the shared secret. A verifier that trusts the header alg
// would accept it. RBACMiddleware pins the family to RSA, so it must not.
func TestSecurity_AlgorithmConfusionIsRejected(t *testing.T) {
	engine, key := buildTestRouter(t, true)

	publicDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	claims := entity.JWTClaims{
		UserID: "attacker",
		Role:   entity.RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	forged, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(publicDER)
	if err != nil {
		t.Fatalf("sign HS256 token: %v", err)
	}

	w := call(t, engine, http.MethodGet, "/api/v1/admin/users", forged)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 for an HS256 token forged from the public key", w.Code)
	}
}

// The "none" algorithm strips the signature entirely.
func TestSecurity_UnsignedTokenIsRejected(t *testing.T) {
	engine, _ := buildTestRouter(t, true)

	claims := entity.JWTClaims{
		UserID: "attacker",
		Role:   entity.RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none token: %v", err)
	}

	w := call(t, engine, http.MethodGet, "/api/v1/admin/users", unsigned)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 for an unsigned token", w.Code)
	}
}

// A tampered payload invalidates the signature.
func TestSecurity_TamperedPayloadIsRejected(t *testing.T) {
	engine, key := buildTestRouter(t, true)
	token := signToken(t, key, entity.RoleUser, time.Now().Add(time.Hour))

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("token has %d parts, want 3", len(parts))
	}
	// Flip a character of the payload segment.
	payload := []byte(parts[1])
	payload[0] ^= 'A' ^ 'B'
	tampered := parts[0] + "." + string(payload) + "." + parts[2]

	w := call(t, engine, http.MethodGet, "/api/v1/playlists", tampered)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 for a tampered payload", w.Code)
	}
}

// ─────────────────────────────────────────────────────────────
// Authorisation — privilege escalation
// ─────────────────────────────────────────────────────────────

func TestSecurity_RoleEscalationIsRefused(t *testing.T) {
	engine, key := buildTestRouter(t, true)

	// A role that does not satisfy the route must get 403, never 401: the caller
	// is authenticated, just not allowed.
	escalations := []struct {
		name   string
		role   entity.UserRole
		method string
		path   string
	}{
		{name: "user reaching a broadcaster route", role: entity.RoleUser, method: http.MethodGet, path: "/api/v1/streams/mine"},
		{name: "user creating a stream", role: entity.RoleUser, method: http.MethodPost, path: "/api/v1/streams"},
		{name: "user reaching the admin panel", role: entity.RoleUser, method: http.MethodGet, path: "/api/v1/admin/users"},
		{name: "user reading admin stats", role: entity.RoleUser, method: http.MethodGet, path: "/api/v1/admin/stats"},
		{name: "broadcaster reaching the admin panel", role: entity.RoleDiffuseur, method: http.MethodGet, path: "/api/v1/admin/users"},
		{name: "broadcaster deleting a user", role: entity.RoleDiffuseur, method: http.MethodDelete, path: "/api/v1/admin/users/u1"},
		{name: "anon role on a user route", role: entity.RoleAnon, method: http.MethodGet, path: "/api/v1/playlists"},
	}

	for _, tt := range escalations {
		t.Run(tt.name, func(t *testing.T) {
			token := signToken(t, key, tt.role, time.Now().Add(time.Hour))
			w := call(t, engine, tt.method, tt.path, token)
			if w.Code != http.StatusForbidden {
				t.Errorf("status = %d, want 403", w.Code)
			}
		})
	}
}

// The audio ingest plane carries the broadcast itself and must not be reachable
// by a plain listener.
func TestSecurity_MediaPlaneRequiresBroadcaster(t *testing.T) {
	engine, key := buildTestRouter(t, true)
	token := signToken(t, key, entity.RoleUser, time.Now().Add(time.Hour))

	for _, route := range []struct{ method, path string }{
		{http.MethodPost, "/api/v1/streams/s1/push"},
		{http.MethodPut, "/api/v1/streams/s1/audio"},
	} {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			if w := call(t, engine, route.method, route.path, token); w.Code != http.StatusForbidden {
				t.Errorf("status = %d with a listener token, want 403", w.Code)
			}
			if w := call(t, engine, route.method, route.path, ""); w.Code != http.StatusUnauthorized {
				t.Errorf("status = %d without a token, want 401", w.Code)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// Operational endpoints
// ─────────────────────────────────────────────────────────────

func TestSecurity_MetricsRequireTheBearerToken(t *testing.T) {
	engine, _ := buildTestRouter(t, true)

	tests := []struct {
		name   string
		header string
		want   int
	}{
		{name: "no header", header: "", want: http.StatusUnauthorized},
		{name: "wrong token", header: "Bearer wrong-token", want: http.StatusUnauthorized},
		{name: "empty bearer", header: "Bearer ", want: http.StatusUnauthorized},
		{name: "correct token", header: "Bearer secret-token", want: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}

// A JWT must not be usable in place of the metrics bearer token.
func TestSecurity_AdminTokenDoesNotOpenMetrics(t *testing.T) {
	engine, key := buildTestRouter(t, true)
	admin := signToken(t, key, entity.RoleAdmin, time.Now().Add(time.Hour))

	w := call(t, engine, http.MethodGet, "/metrics", admin)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401: /metrics uses its own shared secret", w.Code)
	}
}

// ─────────────────────────────────────────────────────────────
// Response hardening
// ─────────────────────────────────────────────────────────────

func TestSecurity_ResponsesCarryHardeningHeaders(t *testing.T) {
	engine, _ := buildTestRouter(t, true)

	w := call(t, engine, http.MethodGet, "/health", "")
	if w.Code != http.StatusOK {
		t.Fatalf("/health status = %d, want 200", w.Code)
	}

	required := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Referrer-Policy":           "no-referrer",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
	}
	for header, want := range required {
		if got := w.Header().Get(header); got != want {
			t.Errorf("header %s = %q, want %q", header, got, want)
		}
	}

	csp := w.Header().Get("Content-Security-Policy")
	for _, directive := range []string{"default-src 'self'", "frame-ancestors 'none'", "object-src 'none'"} {
		if !strings.Contains(csp, directive) {
			t.Errorf("Content-Security-Policy is missing %q: %s", directive, csp)
		}
	}
}

// An error must not echo internals back to the caller.
func TestSecurity_RejectionsDoNotLeakInternals(t *testing.T) {
	engine, _ := buildTestRouter(t, true)

	w := call(t, engine, http.MethodGet, "/api/v1/admin/users", "Bearer garbage")
	body := strings.ToLower(w.Body.String())

	for _, leak := range []string{"panic", "goroutine", "sql", "pgx", "password_hash", ".go:"} {
		if strings.Contains(body, leak) {
			t.Errorf("rejection body leaks %q: %s", leak, w.Body.String())
		}
	}
}

// Path parameters must reach the repository as bound values, never as SQL
// fragments. This wires a capturing repository behind the real middleware chain
// and asserts the payload arrives verbatim: if anything unquoted, unescaped or
// concatenated happened on the way, the captured id would differ.
func TestSecurity_InjectionInPathParameterReachesTheRepositoryVerbatim(t *testing.T) {
	payloads := []struct {
		name    string
		encoded string
		decoded string
	}{
		{name: "tautology", encoded: "1%27%20OR%20%271%27%3D%271", decoded: "1' OR '1'='1"},
		{name: "stacked drop", encoded: "%27%3B%20DROP%20TABLE%20users%3B--", decoded: "'; DROP TABLE users;--"},
		{name: "comment terminator", encoded: "u1%27--", decoded: "u1'--"},
	}

	for _, payload := range payloads {
		t.Run(payload.name, func(t *testing.T) {
			var captured string
			adminRepo := &mock.MockAdminRepository{
				GetUserFn: func(_ context.Context, id string) (*entity.User, error) {
					captured = id
					return &entity.User{ID: id, Email: "captured@streampulse.fr"}, nil
				},
			}

			privateKey, pubPath := writeKeyPair(t)
			cfg := &config.Config{
				JWTPublicKeyPath:   pubPath,
				CORSOrigins:        "http://localhost:3000",
				MetricsBearerToken: "secret-token",
			}
			engine := router.NewRouter(cfg,
				handler.NewAuthHandler(usecase.NewAuthUseCase(&mock.MockUserRepository{}, &mock.MockRefreshTokenRepository{}, "", time.Hour, 720*time.Hour)),
				handler.NewStreamHandler(usecase.NewStreamUseCase(&mock.MockStreamRepository{}, nil)),
				handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(&mock.MockPlaylistRepository{})),
				handler.NewAdminHandler(usecase.NewAdminUseCase(adminRepo)),
				handler.NewFavoriteHandler(usecase.NewFavoriteUseCase(&mock.MockFavoriteRepository{})),
				handler.NewRecommendationHandler(usecase.NewRecommendationUseCase(&mock.MockRecommendationRepository{})),
			)

			token := signToken(t, privateKey, entity.RoleAdmin, time.Now().Add(time.Hour))
			w := call(t, engine, http.MethodGet, "/api/v1/admin/users/"+payload.encoded, token)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (body %s)", w.Code, w.Body.String())
			}
			if captured != payload.decoded {
				t.Errorf("repository received %q, want the payload unchanged: %q", captured, payload.decoded)
			}
			if body := strings.ToLower(w.Body.String()); strings.Contains(body, "syntax") || strings.Contains(body, "sql") {
				t.Errorf("response exposes a database error: %s", w.Body.String())
			}
		})
	}
}
