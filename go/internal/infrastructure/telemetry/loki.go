package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// LokiWriter ships zerolog entries to Grafana Loki without applying network
// backpressure to API requests. Entries are queued, batched and sent by one
// background worker.
type LokiWriter struct {
	endpoint    string
	username    string
	password    string
	serviceName string
	env         string
	httpClient  *http.Client
	queue       chan lokiEntry
	mu          sync.RWMutex
	closed      bool
	workers     sync.WaitGroup
}

const (
	lokiQueueCapacity = 4096
	lokiBatchSize     = 512
	lokiFlushInterval = 250 * time.Millisecond
)

type lokiEntry struct {
	level     string
	timestamp string
	line      string
}

// lokiPushPayload matches the Loki HTTP push API body.
type lokiPushPayload struct {
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"` // [timestamp_ns, log_line]
}

// NewLokiWriter creates a zerolog-compatible asynchronous writer.
// lokiURL example: "https://logs-prod-gb-south-1.grafana.net".
func NewLokiWriter(lokiURL, username, password, serviceName, env string) (*LokiWriter, error) {
	if lokiURL == "" || username == "" || password == "" {
		return nil, fmt.Errorf("loki: missing required config (url, username, password)")
	}
	w := &LokiWriter{
		endpoint:    strings.TrimRight(lokiURL, "/") + "/loki/api/v1/push",
		username:    username,
		password:    password,
		serviceName: serviceName,
		env:         env,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		queue:       make(chan lokiEntry, lokiQueueCapacity),
	}
	w.workers.Add(1)
	go w.run()
	return w, nil
}

// WriteLevel implements zerolog.LevelWriter.
func (w *LokiWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	return w.enqueue(level.String(), p)
}

// Write implements io.Writer when the level is unknown.
func (w *LokiWriter) Write(p []byte) (int, error) {
	return w.enqueue("info", p)
}

// Close drains queued entries. It is idempotent and should run after the HTTP
// server stops accepting requests during graceful shutdown.
func (w *LokiWriter) Close() {
	w.mu.Lock()
	if !w.closed {
		w.closed = true
		close(w.queue)
	}
	w.mu.Unlock()
	w.workers.Wait()
}

func (w *LokiWriter) enqueue(level string, p []byte) (int, error) {
	entry := lokiEntry{
		level:     level,
		timestamp: strconv.FormatInt(time.Now().UnixNano(), 10),
		line:      string(append([]byte(nil), p...)),
	}

	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.closed {
		return len(p), nil
	}
	select {
	case w.queue <- entry:
	default:
		// Best effort: a full bounded queue must never block API or audio I/O.
	}
	return len(p), nil
}

func (w *LokiWriter) run() {
	defer w.workers.Done()
	ticker := time.NewTicker(lokiFlushInterval)
	defer ticker.Stop()

	batch := make([]lokiEntry, 0, lokiBatchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		_ = w.push(batch)
		batch = batch[:0]
	}

	for {
		select {
		case entry, ok := <-w.queue:
			if !ok {
				flush()
				return
			}
			batch = append(batch, entry)
			if len(batch) == lokiBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (w *LokiWriter) push(entries []lokiEntry) error {
	byLevel := make(map[string][][2]string)
	for _, entry := range entries {
		byLevel[entry.level] = append(byLevel[entry.level], [2]string{entry.timestamp, entry.line})
	}

	payload := lokiPushPayload{Streams: make([]lokiStream, 0, len(byLevel))}
	for level, values := range byLevel {
		payload.Streams = append(payload.Streams, lokiStream{
			Stream: map[string]string{
				"service": w.serviceName,
				"env":     w.env,
				"level":   level,
			},
			Values: values,
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("loki: marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, w.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("loki: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(w.username, w.password)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil
	}
	statusCode := resp.StatusCode
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("loki: close response body: %w", err)
	}
	if statusCode/100 != 2 {
		return fmt.Errorf("loki: unexpected status %d", statusCode)
	}
	return nil
}
