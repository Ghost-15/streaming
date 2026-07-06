package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRequiresMetricsTokenInProduction(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("METRICS_BEARER_TOKEN", "")
	t.Setenv("METRICS_BEARER_TOKEN_FILE", "")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "METRICS_BEARER_TOKEN") {
		t.Fatalf("Load() error = %v, want missing metrics token error", err)
	}
}

func TestLoadReadsMetricsTokenFile(t *testing.T) {
	setRequiredEnv(t)
	tokenPath := filepath.Join(t.TempDir(), "metrics_token")
	if err := os.WriteFile(tokenPath, []byte(" token-from-file \n"), 0o600); err != nil {
		t.Fatalf("write token file: %v", err)
	}
	t.Setenv("APP_ENV", "production")
	t.Setenv("METRICS_BEARER_TOKEN", "")
	t.Setenv("METRICS_BEARER_TOKEN_FILE", tokenPath)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MetricsBearerToken != "token-from-file" {
		t.Fatalf("MetricsBearerToken = %q, want token-from-file", cfg.MetricsBearerToken)
	}
}

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SUPABASE_DB_URL", "postgres://example")
	t.Setenv("JWT_PRIVATE_KEY_PATH", "private.pem")
	t.Setenv("JWT_PUBLIC_KEY_PATH", "public.pem")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
}
