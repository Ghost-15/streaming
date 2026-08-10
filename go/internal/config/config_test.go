package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFindsAncestorDotEnvAndResolvesSecretPaths(t *testing.T) {
	projectDir := t.TempDir()
	serverDir := filepath.Join(projectDir, "go", "cmd", "server")
	secretsDir := filepath.Join(projectDir, "secrets")
	if err := os.MkdirAll(serverDir, 0o755); err != nil {
		t.Fatalf("create server directory: %v", err)
	}
	if err := os.MkdirAll(secretsDir, 0o755); err != nil {
		t.Fatalf("create secrets directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "go", "go.mod"), []byte("module example.com/test\n"), 0o600); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	for _, name := range []string{"private.pem", "public.pem", "metrics_token"} {
		if err := os.WriteFile(filepath.Join(secretsDir, name), []byte("test-value"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	dotEnv := strings.Join([]string{
		"APP_ENV=development",
		"SUPABASE_DB_URL=postgres://example",
		// Support both the repository-relative .env.example paths and the
		// Go-module-relative paths used by existing local environments.
		"JWT_PRIVATE_KEY_PATH=../secrets/private.pem",
		"JWT_PUBLIC_KEY_PATH=./secrets/public.pem",
		"METRICS_BEARER_TOKEN_FILE=../secrets/metrics_token",
		"CORS_ALLOWED_ORIGINS=http://localhost:3000",
	}, "\n")
	if err := os.WriteFile(filepath.Join(projectDir, ".env"), []byte(dotEnv), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	unsetEnv(t,
		"APP_ENV",
		"SUPABASE_DB_URL",
		"JWT_PRIVATE_KEY_PATH",
		"JWT_PUBLIC_KEY_PATH",
		"METRICS_BEARER_TOKEN",
		"METRICS_BEARER_TOKEN_FILE",
		"CORS_ALLOWED_ORIGINS",
	)
	changeWorkingDirectory(t, serverDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.JWTPrivateKeyPath != filepath.Join(secretsDir, "private.pem") {
		t.Fatalf("JWTPrivateKeyPath = %q, want project secret path", cfg.JWTPrivateKeyPath)
	}
	if cfg.JWTPublicKeyPath != filepath.Join(secretsDir, "public.pem") {
		t.Fatalf("JWTPublicKeyPath = %q, want project secret path", cfg.JWTPublicKeyPath)
	}
	if cfg.MetricsBearerFile != filepath.Join(secretsDir, "metrics_token") {
		t.Fatalf("MetricsBearerFile = %q, want project secret path", cfg.MetricsBearerFile)
	}
	if cfg.MetricsBearerToken != "test-value" {
		t.Fatalf("MetricsBearerToken = %q, want test-value", cfg.MetricsBearerToken)
	}
}

func TestLoadReportsAllMissingRequiredVariables(t *testing.T) {
	unsetEnv(t,
		"SUPABASE_DB_URL",
		"JWT_PRIVATE_KEY_PATH",
		"JWT_PUBLIC_KEY_PATH",
		"CORS_ALLOWED_ORIGINS",
	)
	changeWorkingDirectory(t, t.TempDir())

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want missing required variables")
	}
	want := "config: missing required env vars: SUPABASE_DB_URL, JWT_PRIVATE_KEY_PATH, JWT_PUBLIC_KEY_PATH, CORS_ALLOWED_ORIGINS"
	if err.Error() != want {
		t.Fatalf("Load() error = %q, want %q", err, want)
	}
}

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

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		value, existed := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
		t.Cleanup(func() {
			if existed {
				_ = os.Setenv(key, value)
				return
			}
			_ = os.Unsetenv(key)
		})
	}
}

func changeWorkingDirectory(t *testing.T, dir string) {
	t.Helper()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousDir)
	})
}
