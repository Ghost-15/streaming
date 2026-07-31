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

func TestLoadValidatesStreamingAndPprofConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr string
	}{
		{name: "invalid duration", key: "STREAM_IDLE_TIMEOUT", value: "later", wantErr: "STREAM_IDLE_TIMEOUT"},
		{name: "invalid chunk size", key: "STREAM_CHUNK_SIZE", value: "100", wantErr: "streaming limits"},
		{name: "invalid boolean", key: "PPROF_ENABLED", value: "perhaps", wantErr: "PPROF_ENABLED"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv(tt.key, tt.value)
			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Load() error = %v, want %q", err, tt.wantErr)
			}
		})
	}

	t.Run("production pprof accepts loopback", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("APP_ENV", "production")
		t.Setenv("METRICS_BEARER_TOKEN", "token")
		t.Setenv("PPROF_ENABLED", "true")
		t.Setenv("PPROF_ADDR", "127.0.0.1:6060")
		if _, err := Load(); err != nil {
			t.Fatalf("Load() error = %v", err)
		}
	})

	t.Run("production pprof rejects public bind", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("APP_ENV", "production")
		t.Setenv("METRICS_BEARER_TOKEN", "token")
		t.Setenv("PPROF_ENABLED", "true")
		t.Setenv("PPROF_ADDR", "0.0.0.0:6060")
		_, err := Load()
		if err == nil || !strings.Contains(err.Error(), "loopback") {
			t.Fatalf("Load() error = %v, want loopback error", err)
		}
	})
}

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SUPABASE_DB_URL", "postgres://example")
	t.Setenv("JWT_PRIVATE_KEY_PATH", "private.pem")
	t.Setenv("JWT_PUBLIC_KEY_PATH", "public.pem")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
}
