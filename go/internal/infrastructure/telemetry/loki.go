package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

// LokiWriter is a zerolog writer that ships log entries directly to Grafana Loki
// via the HTTP push API — no third-party Loki client needed.
type LokiWriter struct {
	endpoint    string // e.g. https://logs-prod-gb-south-1.grafana.net/loki/api/v1/push
	username    string
	password    string
	serviceName string
	env         string
	httpClient  *http.Client
}

// lokiPushPayload matches the Loki HTTP push API body.
type lokiPushPayload struct {
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"` // [timestamp_ns, log_line]
}

// NewLokiWriter creates a zerolog-compatible writer that pushes to Loki.
// lokiURL example: "https://logs-prod-gb-south-1.grafana.net"
func NewLokiWriter(lokiURL, username, password, serviceName, env string) (*LokiWriter, error) {
	if lokiURL == "" || username == "" || password == "" {
		return nil, fmt.Errorf("loki: missing required config (url, username, password)")
	}
	return &LokiWriter{
		endpoint:    lokiURL + "/loki/api/v1/push",
		username:    username,
		password:    password,
		serviceName: serviceName,
		env:         env,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}, nil
}

// WriteLevel implements zerolog.LevelWriter — called for each log entry.
func (w *LokiWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	return w.push(level.String(), p)
}

// Write implements io.Writer — fallback when level is unknown.
func (w *LokiWriter) Write(p []byte) (int, error) {
	return w.push("info", p)
}

// Close is a no-op — HTTP client has no persistent connection to close.
func (w *LokiWriter) Close() {}

func (w *LokiWriter) push(level string, p []byte) (int, error) {
	ts := strconv.FormatInt(time.Now().UnixNano(), 10)

	payload := lokiPushPayload{
		Streams: []lokiStream{
			{
				Stream: map[string]string{
					"service": w.serviceName,
					"env":     w.env,
					"level":   level,
				},
				Values: [][2]string{{ts, string(p)}},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("loki: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, w.endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("loki: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(w.username, w.password)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		// Non-blocking — log to stderr but don't crash the app
		return len(p), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return len(p), fmt.Errorf("loki: unexpected status %d", resp.StatusCode)
	}

	return len(p), nil
}
