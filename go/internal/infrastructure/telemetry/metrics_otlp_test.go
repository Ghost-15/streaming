package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestParseOTLPHTTPEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		signal   string
		want     string
	}{
		{name: "base path", endpoint: "https://example.com/otlp", signal: "metrics", want: "https://example.com/otlp/v1/metrics"},
		{name: "trailing slash", endpoint: "https://example.com/otlp/", signal: "traces", want: "https://example.com/otlp/v1/traces"},
		{name: "root", endpoint: "http://localhost:4318", signal: "metrics", want: "http://localhost:4318/v1/metrics"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOTLPHTTPEndpoint(tt.endpoint, tt.signal)
			if err != nil {
				t.Fatalf("parseOTLPHTTPEndpoint() error = %v", err)
			}
			if got.String() != tt.want {
				t.Fatalf("parseOTLPHTTPEndpoint() = %q, want %q", got, tt.want)
			}
		})
	}
	if _, err := parseOTLPHTTPEndpoint("collector:4318", "metrics"); err == nil {
		t.Fatal("parseOTLPHTTPEndpoint() error = nil for a non-HTTP endpoint")
	}
}

func TestObserveAPIRequestDurationRecordsOTLPMetric(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	instruments, err := newOTelMetricInstruments(provider.Meter("test"), "streampulse-api", "production")
	if err != nil {
		t.Fatalf("newOTelMetricInstruments() error = %v", err)
	}
	directMetrics.Store(instruments)
	t.Cleanup(func() { directMetrics.Store(nil) })

	ObserveAPIRequestDuration(context.Background(), "/health", "GET", "200", 0.125)

	var metrics metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &metrics); err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	for _, scope := range metrics.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if metric.Name != "streampulse_api_request_duration_seconds" {
				continue
			}
			histogram, ok := metric.Data.(metricdata.Histogram[float64])
			if !ok || len(histogram.DataPoints) != 1 {
				t.Fatalf("metric data = %#v, want one float64 histogram data point", metric.Data)
			}
			point := histogram.DataPoints[0]
			if point.Count != 1 || point.Sum != 0.125 {
				t.Fatalf("histogram count/sum = %d/%v, want 1/0.125", point.Count, point.Sum)
			}
			for key, want := range map[attribute.Key]string{
				"service": "streampulse-api", "env": "production", "route": "/health", "method": "GET", "status": "200",
			} {
				value, found := point.Attributes.Value(key)
				if !found || value.AsString() != want {
					t.Fatalf("attribute %q = %q (found %v), want %q", key, value.AsString(), found, want)
				}
			}
			return
		}
	}
	t.Fatal("OTLP API request metric was not collected")
}

func TestMetricHelpersRecordEveryOTLPInstrument(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	instruments, err := newOTelMetricInstruments(provider.Meter("test"), "streampulse-api", "production")
	if err != nil {
		t.Fatalf("newOTelMetricInstruments() error = %v", err)
	}
	directMetrics.Store(instruments)
	t.Cleanup(func() { directMetrics.Store(nil) })

	ActiveStreams.Set(0)
	OnlineUsers.Set(0)
	ListenersPerStream.Reset()
	IncrementActiveStreams()
	DecrementActiveStreams()
	IncrementOnlineUsers()
	DecrementOnlineUsers()
	IncrementListeners("stream-1")
	DecrementListeners("stream-1")
	RecordStreamStart()
	RecordListenerDisconnect()
	AddAudioIngestBytes("stream-1", 1024)
	AddAudioEgressBytes("stream-1", 768)
	RecordAudioChunk("stream-1", "ingest")
	AddDroppedAudioChunks("stream-1", 2)
	SetAudioBroadcaster("stream-1", true)
	SetAudioBroadcaster("stream-1", false)
	ObserveAudioChunkSize(1024)
	ObserveListenerSessionDuration(3.5)
	ObserveBroadcasterSessionDuration(4.5)
	ObserveAPIRequestDuration(context.Background(), "/health", "GET", "200", 0.05)

	var metrics metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &metrics); err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	want := map[string]bool{
		"streampulse_active_streams":                             false,
		"streampulse_online_users":                               false,
		"streampulse_listeners_per_stream":                       false,
		"streampulse_stream_start_total":                         false,
		"streampulse_listener_disconnect_total":                  false,
		"streampulse_audio_ingest_bytes_total":                   false,
		"streampulse_audio_egress_bytes_total":                   false,
		"streampulse_audio_chunks_total":                         false,
		"streampulse_audio_dropped_chunks_total":                 false,
		"streampulse_audio_broadcasters":                         false,
		"streampulse_audio_chunk_size_bytes":                     false,
		"streampulse_audio_listener_session_duration_seconds":    false,
		"streampulse_audio_broadcaster_session_duration_seconds": false,
		"streampulse_api_request_duration_seconds":               false,
	}
	for _, scope := range metrics.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if _, expected := want[metric.Name]; expected {
				want[metric.Name] = true
			}
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("OTLP metric %q was not collected", name)
		}
	}
}

func TestInitMetricsExportsHTTPProtobuf(t *testing.T) {
	requestPath := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath <- r.URL.Path
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "")
	shutdown, err := InitMetrics(context.Background(), "streampulse-api", "test", "production", server.URL+"/otlp/")
	if err != nil {
		t.Fatalf("InitMetrics() error = %v", err)
	}
	t.Cleanup(func() {
		if directMetrics.Load() != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = shutdown(ctx)
		}
	})

	RecordStreamStart()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		t.Fatalf("metrics shutdown/export error = %v", err)
	}

	select {
	case got := <-requestPath:
		if got != "/otlp/v1/metrics" {
			t.Fatalf("metric export path = %q, want /otlp/v1/metrics", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for OTLP metric export")
	}
}

func TestInitMetricsRejectsEmptyEndpoint(t *testing.T) {
	if _, err := InitMetrics(context.Background(), "svc", "ns", "dev", ""); err == nil {
		t.Fatal("InitMetrics() error = nil for an empty endpoint")
	}
}
