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

	// OnlineUsers tracks the number of unique users currently connected to streams.
	OnlineUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "streampulse_online_users",
		Help: "Number of unique users currently connected to a stream.",
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

	// ListenerDisconnectTotal counts listener disconnects from live streams.
	ListenerDisconnectTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "streampulse_listener_disconnect_total",
		Help: "Total number of listener disconnects since process boot.",
	})

	// AudioIngestBytesTotal and AudioEgressBytesTotal are fed only by bytes
	// that cross the live-audio endpoints. They are not inferred from database
	// join/leave events.
	AudioIngestBytesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "streampulse_audio_ingest_bytes_total",
		Help: "Audio bytes received from broadcasters.",
	}, []string{"stream_id"})

	AudioEgressBytesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "streampulse_audio_egress_bytes_total",
		Help: "Audio bytes written successfully to listener HTTP responses.",
	}, []string{"stream_id"})

	AudioChunksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "streampulse_audio_chunks_total",
		Help: "Audio chunks handled by direction.",
	}, []string{"stream_id", "direction"})

	AudioDroppedChunksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "streampulse_audio_dropped_chunks_total",
		Help: "Audio chunks dropped because a listener buffer was full.",
	}, []string{"stream_id"})

	AudioBroadcasters = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "streampulse_audio_broadcasters",
		Help: "Broadcasters currently ingesting audio (zero or one per stream).",
	}, []string{"stream_id"})

	AudioChunkSizeBytes = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "streampulse_audio_chunk_size_bytes",
		Help:    "Size of audio chunks received from broadcasters.",
		Buckets: prometheus.ExponentialBuckets(256, 2, 8),
	})

	ListenerSessionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "streampulse_audio_listener_session_duration_seconds",
		Help:    "Duration of real HTTP audio listener connections.",
		Buckets: []float64{1, 5, 15, 30, 60, 300, 900, 3600},
	})

	BroadcasterSessionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "streampulse_audio_broadcaster_session_duration_seconds",
		Help:    "Duration of real HTTP audio ingestion connections.",
		Buckets: []float64{1, 5, 15, 30, 60, 300, 900, 3600},
	})

	// APIRequestDuration measures the duration of HTTP requests (histogram).
	// Labels: route (Gin pattern, e.g. /api/v1/playlists/:id), method, status.
	APIRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "streampulse_api_request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"route", "method", "status"})
)
