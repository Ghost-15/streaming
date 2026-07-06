package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration loaded from environment variables.
// 12-Factor App: no hardcoded values — everything is ENV.
type Config struct {
	Port                 string
	SupabaseDBURL        string
	JWTPrivateKeyPath    string
	JWTPublicKeyPath     string
	OTELEndpoint         string
	OTELServiceNamespace string
	OTELDeploymentEnv    string
	CORSOrigins          string
	MetricsBearerToken   string
	MetricsBearerFile    string
	Env                  string // "development" | "production"
}

// Load reads configuration from environment variables.
// In development, it also loads a .env file if present.
// Returns an error if any required variable is missing.
func Load() (*Config, error) {
	// In development, load .env file (ignored if missing in production).
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	metricsBearerFile := os.Getenv("METRICS_BEARER_TOKEN_FILE")
	metricsBearerToken, err := readSecretValue(metricsBearerFile, os.Getenv("METRICS_BEARER_TOKEN"))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Port:                 getEnv("PORT", "8080"),
		SupabaseDBURL:        os.Getenv("SUPABASE_DB_URL"),
		JWTPrivateKeyPath:    os.Getenv("JWT_PRIVATE_KEY_PATH"),
		JWTPublicKeyPath:     os.Getenv("JWT_PUBLIC_KEY_PATH"),
		OTELEndpoint:         getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		OTELServiceNamespace: getEnv("OTEL_SERVICE_NAMESPACE", "my-application-group"),
		OTELDeploymentEnv:    getEnv("OTEL_DEPLOYMENT_ENVIRONMENT", "production"),
		CORSOrigins:          os.Getenv("CORS_ALLOWED_ORIGINS"),
		MetricsBearerToken:   metricsBearerToken,
		MetricsBearerFile:    metricsBearerFile,
		Env:                  getEnv("APP_ENV", "development"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"SUPABASE_DB_URL":      c.SupabaseDBURL,
		"JWT_PRIVATE_KEY_PATH": c.JWTPrivateKeyPath,
		"JWT_PUBLIC_KEY_PATH":  c.JWTPublicKeyPath,
		"CORS_ALLOWED_ORIGINS": c.CORSOrigins,
	}
	for key, val := range required {
		if val == "" {
			return fmt.Errorf("config: missing required env var %s", key)
		}
	}
	if c.Env == "production" && c.MetricsBearerToken == "" {
		return fmt.Errorf("config: missing METRICS_BEARER_TOKEN or METRICS_BEARER_TOKEN_FILE in production")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func readSecretValue(filePath, fallback string) (string, error) {
	if filePath == "" {
		return strings.TrimSpace(fallback), nil
	}
	value, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("config: read %s: %w", filePath, err)
	}
	return strings.TrimSpace(string(value)), nil
}
