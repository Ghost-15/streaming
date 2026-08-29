package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// operationalRoutes are served by the process itself rather than by the REST
// API: /metrics is the Prometheus scrape target and /swagger renders the
// contract. Neither belongs to the documented resource surface.
var operationalRoutes = map[string]bool{
	"GET /metrics":      true,
	"GET /swagger/*any": true,
}

type openAPISpec struct {
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

func loadSpec(t *testing.T) openAPISpec {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "openapi", "swagger.json"))
	if err != nil {
		t.Fatalf("read generated spec: %v (run: swag init --dir ./cmd/server,./internal/handler,./internal/entity --generalInfo main.go --output docs/openapi --parseInternal --parseDepth 2)", err)
	}

	var spec openAPISpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("decode generated spec: %v", err)
	}
	if len(spec.Paths) == 0 {
		t.Fatal("generated spec declares no path")
	}
	return spec
}

// ginPathToOpenAPI rewrites Gin wildcards into the OpenAPI template syntax:
// /streams/:id/stop becomes /streams/{id}/stop.
func ginPathToOpenAPI(path string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			segments[i] = "{" + segment[1:] + "}"
		}
	}
	return strings.Join(segments, "/")
}

// TestOpenAPISpecCoversEveryRoute is the guard against documentation drift: the
// committed spec is generated from annotations, so a handler added without them
// would silently disappear from the contract. This fails instead.
func TestOpenAPISpecCoversEveryRoute(t *testing.T) {
	spec := loadSpec(t)
	engine := newTestRouter(t)

	var undocumented []string
	for _, route := range engine.Routes() {
		key := route.Method + " " + route.Path
		if operationalRoutes[key] {
			continue
		}

		operations, ok := spec.Paths[ginPathToOpenAPI(route.Path)]
		if !ok {
			undocumented = append(undocumented, key)
			continue
		}
		if _, ok := operations[strings.ToLower(route.Method)]; !ok {
			undocumented = append(undocumented, key)
		}
	}

	if len(undocumented) > 0 {
		t.Errorf("routes missing from the OpenAPI spec:\n  %s\nAdd swaggo annotations to the handler, then regenerate the spec.",
			strings.Join(undocumented, "\n  "))
	}
}

// TestOpenAPISpecHasNoStaleOperation catches the opposite drift: an operation
// documented in the spec that no longer maps to a registered route.
func TestOpenAPISpecHasNoStaleOperation(t *testing.T) {
	spec := loadSpec(t)
	engine := newTestRouter(t)

	registered := make(map[string]bool)
	for _, route := range engine.Routes() {
		registered[strings.ToLower(route.Method)+" "+ginPathToOpenAPI(route.Path)] = true
	}

	var stale []string
	for path, operations := range spec.Paths {
		for method := range operations {
			if !registered[method+" "+path] {
				stale = append(stale, strings.ToUpper(method)+" "+path)
			}
		}
	}

	if len(stale) > 0 {
		t.Errorf("operations documented but not routed:\n  %s", strings.Join(stale, "\n  "))
	}
}

// TestSwaggerUIIsServed proves the contract is reachable at runtime, not just
// generated at build time.
func TestSwaggerUIIsServed(t *testing.T) {
	engine := newTestRouter(t)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("/swagger/index.html status = %d, want 200", w.Code)
	}
	if body := w.Body.String(); !strings.Contains(strings.ToLower(body), "swagger") {
		t.Error("/swagger/index.html did not render the Swagger UI document")
	}
}

// TestSwaggerUICanBeDisabled covers the 12-Factor switch: an operator must be
// able to remove the documentation surface without rebuilding the image.
func TestSwaggerUICanBeDisabled(t *testing.T) {
	engine := newTestRouterWithSwagger(t, false)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("/swagger/index.html status = %d with SWAGGER_ENABLED=false, want 404", w.Code)
	}
}
