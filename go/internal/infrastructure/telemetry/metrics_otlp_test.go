package telemetry

import (
	"context"
	"testing"

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
