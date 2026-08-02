package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration loaded from environment variables.
// 12-Factor App: no hardcoded values — everything is ENV.
type Config struct {
	Port                  string
	SupabaseDBURL         string
	JWTPrivateKeyPath     string
	JWTPublicKeyPath      string
	OTELEndpoint          string
	OTELServiceNamespace  string
	OTELDeploymentEnv     string
	CORSOrigins           string
	MetricsBearerToken    string
	MetricsBearerFile     string
	Env                   string // "development" | "production"
	HTTPReadHeaderTimeout time.Duration
	HTTPIdleTimeout       time.Duration
	ShutdownTimeout       time.Duration
	StreamMaxDuration     time.Duration
	StreamIdleTimeout     time.Duration
	StreamWriteTimeout    time.Duration
	StreamMaxIngestBytes  int64
	StreamChunkSize       int
	StreamClientBuffer    int
	PprofEnabled          bool
	PprofAddr             string
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

	httpReadHeaderTimeout, err := durationEnv("HTTP_READ_HEADER_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, err
	}
	httpIdleTimeout, err := durationEnv("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return nil, err
	}
	shutdownTimeout, err := durationEnv("SHUTDOWN_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, err
	}
	streamMaxDuration, err := durationEnv("STREAM_MAX_DURATION", 6*time.Hour)
	if err != nil {
		return nil, err
	}
	streamIdleTimeout, err := durationEnv("STREAM_IDLE_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, err
	}
	streamWriteTimeout, err := durationEnv("STREAM_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, err
	}
	streamMaxIngestBytes, err := int64Env("STREAM_MAX_INGEST_BYTES", 8<<30)
	if err != nil {
		return nil, err
	}
	streamChunkSize, err := intEnv("STREAM_CHUNK_SIZE", 32<<10)
	if err != nil {
		return nil, err
	}
	streamClientBuffer, err := intEnv("STREAM_CLIENT_BUFFER", 64)
	if err != nil {
		return nil, err
	}
	pprofEnabled, err := boolEnv("PPROF_ENABLED", false)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Port:                  getEnv("PORT", "8080"),
		SupabaseDBURL:         os.Getenv("SUPABASE_DB_URL"),
		JWTPrivateKeyPath:     os.Getenv("JWT_PRIVATE_KEY_PATH"),
		JWTPublicKeyPath:      os.Getenv("JWT_PUBLIC_KEY_PATH"),
		OTELEndpoint:          getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		OTELServiceNamespace:  getEnv("OTEL_SERVICE_NAMESPACE", "my-application-group"),
		OTELDeploymentEnv:     getEnv("OTEL_DEPLOYMENT_ENVIRONMENT", "production"),
		CORSOrigins:           os.Getenv("CORS_ALLOWED_ORIGINS"),
		MetricsBearerToken:    metricsBearerToken,
		MetricsBearerFile:     metricsBearerFile,
		Env:                   getEnv("APP_ENV", "development"),
		HTTPReadHeaderTimeout: httpReadHeaderTimeout,
		HTTPIdleTimeout:       httpIdleTimeout,
		ShutdownTimeout:       shutdownTimeout,
		StreamMaxDuration:     streamMaxDuration,
		StreamIdleTimeout:     streamIdleTimeout,
		StreamWriteTimeout:    streamWriteTimeout,
		StreamMaxIngestBytes:  streamMaxIngestBytes,
		StreamChunkSize:       streamChunkSize,
		StreamClientBuffer:    streamClientBuffer,
		PprofEnabled:          pprofEnabled,
		PprofAddr:             getEnv("PPROF_ADDR", "127.0.0.1:6060"),
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
	if c.StreamMaxDuration <= 0 || c.StreamIdleTimeout <= 0 || c.StreamWriteTimeout <= 0 ||
		c.ShutdownTimeout <= 0 || c.HTTPReadHeaderTimeout <= 0 || c.HTTPIdleTimeout <= 0 {
		return fmt.Errorf("config: timeout values must be positive")
	}
	if c.StreamMaxIngestBytes <= 0 || c.StreamChunkSize < 1024 || c.StreamChunkSize > 1<<20 ||
		c.StreamClientBuffer < 1 || c.StreamClientBuffer > 4096 {
		return fmt.Errorf("config: invalid streaming limits")
	}
	if c.Env == "production" && c.PprofEnabled && !isLoopbackAddress(c.PprofAddr) {
		return fmt.Errorf("config: PPROF_ADDR must bind to loopback in production")
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

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := getEnv(key, fallback.String())
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("config: invalid %s: %w", key, err)
	}
	return parsed, nil
}

func intEnv(key string, fallback int) (int, error) {
	value := getEnv(key, strconv.Itoa(fallback))
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("config: invalid %s: %w", key, err)
	}
	return parsed, nil
}

func int64Env(key string, fallback int64) (int64, error) {
	value := getEnv(key, strconv.FormatInt(fallback, 10))
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("config: invalid %s: %w", key, err)
	}
	return parsed, nil
}

func boolEnv(key string, fallback bool) (bool, error) {
	value := getEnv(key, strconv.FormatBool(fallback))
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("config: invalid %s: %w", key, err)
	}
	return parsed, nil
}

func isLoopbackAddress(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
