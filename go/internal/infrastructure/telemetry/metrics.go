// Package telemetry — metrics layer.
// Sprint 3 — US-010. Custom Prometheus collectors for the StreamPulse domain.
package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Custom Prometheus collectors exposed at /metrics.
// All metric names are prefixed "streampulse_" per the StreamPulse naming convention.
var (
	// ActiveStreams tracks the number of currently live streams (gauge).
	// Incremented on StreamUseCase.Start, decremented on StreamUseCase.End.
	ActiveStreams = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "streampulse_active_streams",
		Help: "Number of currently live streams.",
	})

	// ListenersPerStream tracks the number of listeners attached to each stream (gauge).
	// Incremented on Hub.Register, decremented on Hub.Unregister.
	ListenersPerStream = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "streampulse_listeners_per_stream",
		Help: "Number of active listeners per stream.",
	}, []string{"stream_id"})

	// StreamStartTotal counts the total number of streams ever started (counter).
	StreamStartTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "streampulse_stream_start_total",
		Help: "Total number of streams started since process boot.",
	})

	// APIRequestDuration measures the duration of HTTP requests (histogram).
	// Labels: route (Gin pattern, e.g. /api/v1/playlists/:id), method, status.
	APIRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "streampulse_api_request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"route", "method", "status"})
)
