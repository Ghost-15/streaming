package telemetry_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
)

func TestActiveStreams_IncDec(t *testing.T) {
	telemetry.ActiveStreams.Set(0)
	telemetry.ActiveStreams.Inc()
	telemetry.ActiveStreams.Inc()
	if got := testutil.ToFloat64(telemetry.ActiveStreams); got != 2 {
		t.Errorf("ActiveStreams after 2 Inc = %v, want 2", got)
	}
	telemetry.ActiveStreams.Dec()
	if got := testutil.ToFloat64(telemetry.ActiveStreams); got != 1 {
		t.Errorf("ActiveStreams after Dec = %v, want 1", got)
	}
}

func TestStreamStartTotal_Inc(t *testing.T) {
	before := testutil.ToFloat64(telemetry.StreamStartTotal)
	telemetry.StreamStartTotal.Inc()
	telemetry.StreamStartTotal.Inc()
	after := testutil.ToFloat64(telemetry.StreamStartTotal)
	if after-before != 2 {
		t.Errorf("StreamStartTotal delta = %v, want 2", after-before)
	}
}

func TestListenersPerStream_PerLabel(t *testing.T) {
	telemetry.ListenersPerStream.Reset()
	telemetry.ListenersPerStream.WithLabelValues("stream-A").Inc()
	telemetry.ListenersPerStream.WithLabelValues("stream-A").Inc()
	telemetry.ListenersPerStream.WithLabelValues("stream-B").Inc()

	if got := testutil.ToFloat64(telemetry.ListenersPerStream.WithLabelValues("stream-A")); got != 2 {
		t.Errorf("ListenersPerStream[A] = %v, want 2", got)
	}
	if got := testutil.ToFloat64(telemetry.ListenersPerStream.WithLabelValues("stream-B")); got != 1 {
		t.Errorf("ListenersPerStream[B] = %v, want 1", got)
	}

	telemetry.ListenersPerStream.WithLabelValues("stream-A").Dec()
	if got := testutil.ToFloat64(telemetry.ListenersPerStream.WithLabelValues("stream-A")); got != 1 {
		t.Errorf("ListenersPerStream[A] after Dec = %v, want 1", got)
	}
}

func TestAPIRequestDuration_Observe(t *testing.T) {
	telemetry.APIRequestDuration.Reset()
	hist := telemetry.APIRequestDuration.WithLabelValues("/health", "GET", "200")
	hist.Observe(0.123)
	hist.Observe(0.456)

	// CollectAndCount returns the number of samples collected for the matching metric.
	count := testutil.CollectAndCount(telemetry.APIRequestDuration)
	if count != 1 {
		t.Errorf("APIRequestDuration label count = %d, want 1", count)
	}
}
