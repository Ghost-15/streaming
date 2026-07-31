package telemetry_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
)

func TestPprofServerUsesIsolatedExplicitMux(t *testing.T) {
	server := telemetry.NewPprofServer("127.0.0.1:0")
	if server.Addr != "127.0.0.1:0" {
		t.Fatalf("Addr = %q", server.Addr)
	}
	for _, path := range []string{"/debug/pprof/", "/debug/pprof/goroutine?debug=1"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want 200", path, response.Code)
		}
	}
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("unregistered route status = %d, want 404", response.Code)
	}
}
