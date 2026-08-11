package telemetry

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const metricExportInterval = 15 * time.Second

type otelMetricInstruments struct {
	attributes                 []attribute.KeyValue
	activeStreams              otelmetric.Int64UpDownCounter
	onlineUsers                otelmetric.Int64UpDownCounter
	listenersPerStream         otelmetric.Int64UpDownCounter
	streamStartTotal           otelmetric.Int64Counter
	listenerDisconnectTotal    otelmetric.Int64Counter
	audioIngestBytesTotal      otelmetric.Int64Counter
	audioEgressBytesTotal      otelmetric.Int64Counter
	audioChunksTotal           otelmetric.Int64Counter
	audioDroppedChunksTotal    otelmetric.Int64Counter
	audioBroadcasters          otelmetric.Int64Gauge
	audioChunkSizeBytes        otelmetric.Float64Histogram
	listenerSessionDuration    otelmetric.Float64Histogram
	broadcasterSessionDuration otelmetric.Float64Histogram
	apiRequestDuration         otelmetric.Float64Histogram
}

var directMetrics atomic.Pointer[otelMetricInstruments]

// InitMetrics configures direct OTLP metric export. It deliberately runs in
// addition to the Prometheus collectors so local scraping and /metrics remain
// available without requiring a production Alloy worker.
func InitMetrics(
	ctx context.Context,
	serviceName, serviceNamespace, deploymentEnvironment, endpoint string,
) (shutdown func(context.Context) error, err error) {
	if strings.TrimSpace(endpoint) == "" {
		return nil, fmt.Errorf("telemetry: metrics endpoint is empty")
	}

	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	var exporter sdkmetric.Exporter
	if protocol == "http/protobuf" || strings.HasPrefix(endpoint, "https://") {
		u, parseErr := parseOTLPHTTPEndpoint(endpoint, "metrics")
		if parseErr != nil {
			return nil, parseErr
		}
		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(u.Host),
			otlpmetrichttp.WithURLPath(u.Path),
			otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
		}
		if u.Scheme == "http" {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
		exporter, err = otlpmetrichttp.New(ctx, opts...)
	} else {
		exporter, err = otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(trimEndpointScheme(endpoint)),
			otlpmetricgrpc.WithInsecure(),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("telemetry: create metrics exporter: %w", err)
	}

	res, err := newResource(ctx, serviceName, serviceNamespace, deploymentEnvironment)
	if err != nil {
		return nil, err
	}
	reader := sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(metricExportInterval))
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)
	instruments, err := newOTelMetricInstruments(
		mp.Meter("github.com/Ghost-15/streaming"),
		serviceName,
		deploymentEnvironment,
	)
	if err != nil {
		_ = mp.Shutdown(ctx)
		return nil, err
	}

	directMetrics.Store(instruments)
	otel.SetMeterProvider(mp)
	return func(shutdownCtx context.Context) error {
		directMetrics.Store(nil)
		return mp.Shutdown(shutdownCtx)
	}, nil
}

func parseOTLPHTTPEndpoint(endpoint, signal string) (*url.URL, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		if err == nil {
			err = fmt.Errorf("expected an http or https URL")
		}
		return nil, fmt.Errorf("telemetry: invalid endpoint: %w", err)
	}
	u.Path = strings.TrimSuffix(u.Path, "/") + "/v1/" + signal
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u, nil
}

func trimEndpointScheme(endpoint string) string {
	endpoint = strings.TrimPrefix(endpoint, "http://")
	return strings.TrimPrefix(endpoint, "https://")
}

func newOTelMetricInstruments(
	meter otelmetric.Meter,
	serviceName, deploymentEnvironment string,
) (*otelMetricInstruments, error) {
	instruments := &otelMetricInstruments{
		attributes: []attribute.KeyValue{
			attribute.String("service", serviceName),
			attribute.String("env", deploymentEnvironment),
		},
	}
	var err error
	if instruments.activeStreams, err = meter.Int64UpDownCounter(
		"streampulse_active_streams",
		otelmetric.WithDescription("Number of currently live streams."),
	); err != nil {
		return nil, fmt.Errorf("telemetry: create active streams metric: %w", err)
	}
	if instruments.onlineUsers, err = meter.Int64UpDownCounter(
		"streampulse_online_users",
		otelmetric.WithDescription("Number of unique users currently connected to a stream."),
	); err != nil {
		return nil, fmt.Errorf("telemetry: create online users metric: %w", err)
	}
	if instruments.listenersPerStream, err = meter.Int64UpDownCounter(
		"streampulse_listeners_per_stream",
		otelmetric.WithDescription("Number of active listeners per stream."),
	); err != nil {
		return nil, fmt.Errorf("telemetry: create listeners metric: %w", err)
	}
	if instruments.streamStartTotal, err = meter.Int64Counter(
		"streampulse_stream_start_total",
		otelmetric.WithDescription("Total number of streams started since process boot."),
	); err != nil {
		return nil, fmt.Errorf("telemetry: create stream starts metric: %w", err)
	}
	if instruments.listenerDisconnectTotal, err = meter.Int64Counter(
		"streampulse_listener_disconnect_total",
		otelmetric.WithDescription("Total number of listener disconnects since process boot."),
	); err != nil {
		return nil, fmt.Errorf("telemetry: create listener disconnect metric: %w", err)
	}
	if instruments.audioIngestBytesTotal, err = meter.Int64Counter(
		"streampulse_audio_ingest_bytes_total",
		otelmetric.WithDescription("Audio bytes received from broadcasters."),
	); err != nil {
		return nil, fmt.Errorf("telemetry: create audio ingest metric: %w", err)
	}
	if instruments.audioEgressBytesTotal, err = meter.Int64Counter(
		"streampulse_audio_egress_bytes_total",
		otelmetric.WithDescription("Audio bytes written successfully to listener responses."),
	); err != nil {
		return nil, fmt.Errorf("telemetry: create audio egress metric: %w", err)
	}
	if instruments.audioChunksTotal, err = meter.Int64Counter(
		"streampulse_audio_chunks_total",
		otelmetric.WithDescription("Audio chunks handled by direction."),
	); err != nil {
		return nil, fmt.Errorf("telemetry: create audio chunks metric: %w", err)
	}
	if instruments.audioDroppedChunksTotal, err = meter.Int64Counter(
		"streampulse_audio_dropped_chunks_total",
		otelmetric.WithDescription("Audio chunks dropped because a listener buffer was full."),
	); err != nil {
		return nil, fmt.Errorf("telemetry: create dropped audio chunks metric: %w", err)
	}
	if instruments.audioBroadcasters, err = meter.Int64Gauge(
		"streampulse_audio_broadcasters",
		otelmetric.WithDescription("Broadcasters currently ingesting audio."),
	); err != nil {
		return nil, fmt.Errorf("telemetry: create audio broadcasters metric: %w", err)
	}
	if instruments.audioChunkSizeBytes, err = meter.Float64Histogram(
		"streampulse_audio_chunk_size_bytes",
		otelmetric.WithDescription("Size of audio chunks received from broadcasters."),
		otelmetric.WithExplicitBucketBoundaries(256, 512, 1024, 2048, 4096, 8192, 16384, 32768),
	); err != nil {
		return nil, fmt.Errorf("telemetry: create audio chunk size metric: %w", err)
	}
	sessionBuckets := otelmetric.WithExplicitBucketBoundaries(1, 5, 15, 30, 60, 300, 900, 3600)
	if instruments.listenerSessionDuration, err = meter.Float64Histogram(
		"streampulse_audio_listener_session_duration_seconds",
		otelmetric.WithDescription("Duration of real HTTP audio listener connections."),
		sessionBuckets,
	); err != nil {
		return nil, fmt.Errorf("telemetry: create listener session metric: %w", err)
	}
	if instruments.broadcasterSessionDuration, err = meter.Float64Histogram(
		"streampulse_audio_broadcaster_session_duration_seconds",
		otelmetric.WithDescription("Duration of real HTTP audio ingestion connections."),
		sessionBuckets,
	); err != nil {
		return nil, fmt.Errorf("telemetry: create broadcaster session metric: %w", err)
	}
	if instruments.apiRequestDuration, err = meter.Float64Histogram(
		"streampulse_api_request_duration_seconds",
		otelmetric.WithDescription("HTTP request duration in seconds."),
		otelmetric.WithExplicitBucketBoundaries(.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10),
	); err != nil {
		return nil, fmt.Errorf("telemetry: create API request duration metric: %w", err)
	}
	return instruments, nil
}

func (m *otelMetricInstruments) options(attributes ...attribute.KeyValue) otelmetric.MeasurementOption {
	all := make([]attribute.KeyValue, 0, len(m.attributes)+len(attributes))
	all = append(all, m.attributes...)
	all = append(all, attributes...)
	return otelmetric.WithAttributes(all...)
}

func IncrementActiveStreams() {
	ActiveStreams.Inc()
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.activeStreams.Add(context.Background(), 1, metrics.options())
	}
}

func DecrementActiveStreams() {
	ActiveStreams.Dec()
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.activeStreams.Add(context.Background(), -1, metrics.options())
	}
}

func IncrementOnlineUsers() {
	OnlineUsers.Inc()
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.onlineUsers.Add(context.Background(), 1, metrics.options())
	}
}

func DecrementOnlineUsers() {
	OnlineUsers.Dec()
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.onlineUsers.Add(context.Background(), -1, metrics.options())
	}
}

func IncrementListeners(streamID string) {
	ListenersPerStream.WithLabelValues(streamID).Inc()
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.listenersPerStream.Add(context.Background(), 1,
			metrics.options(attribute.String("stream_id", streamID)))
	}
}

func DecrementListeners(streamID string) {
	ListenersPerStream.WithLabelValues(streamID).Dec()
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.listenersPerStream.Add(context.Background(), -1,
			metrics.options(attribute.String("stream_id", streamID)))
	}
}

func RecordStreamStart() {
	StreamStartTotal.Inc()
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.streamStartTotal.Add(context.Background(), 1, metrics.options())
	}
}

func RecordListenerDisconnect() {
	ListenerDisconnectTotal.Inc()
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.listenerDisconnectTotal.Add(context.Background(), 1, metrics.options())
	}
}

func AddAudioIngestBytes(streamID string, count int) {
	AudioIngestBytesTotal.WithLabelValues(streamID).Add(float64(count))
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.audioIngestBytesTotal.Add(context.Background(), int64(count),
			metrics.options(attribute.String("stream_id", streamID)))
	}
}

func AddAudioEgressBytes(streamID string, count int) {
	AudioEgressBytesTotal.WithLabelValues(streamID).Add(float64(count))
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.audioEgressBytesTotal.Add(context.Background(), int64(count),
			metrics.options(attribute.String("stream_id", streamID)))
	}
}

func RecordAudioChunk(streamID, direction string) {
	AudioChunksTotal.WithLabelValues(streamID, direction).Inc()
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.audioChunksTotal.Add(context.Background(), 1, metrics.options(
			attribute.String("stream_id", streamID),
			attribute.String("direction", direction),
		))
	}
}

func AddDroppedAudioChunks(streamID string, count int) {
	AudioDroppedChunksTotal.WithLabelValues(streamID).Add(float64(count))
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.audioDroppedChunksTotal.Add(context.Background(), int64(count),
			metrics.options(attribute.String("stream_id", streamID)))
	}
}

func SetAudioBroadcaster(streamID string, active bool) {
	value := int64(0)
	if active {
		value = 1
	}
	AudioBroadcasters.WithLabelValues(streamID).Set(float64(value))
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.audioBroadcasters.Record(context.Background(), value,
			metrics.options(attribute.String("stream_id", streamID)))
	}
}

func ObserveAudioChunkSize(size int) {
	AudioChunkSizeBytes.Observe(float64(size))
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.audioChunkSizeBytes.Record(context.Background(), float64(size), metrics.options())
	}
}

func ObserveListenerSessionDuration(seconds float64) {
	ListenerSessionDuration.Observe(seconds)
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.listenerSessionDuration.Record(context.Background(), seconds, metrics.options())
	}
}

func ObserveBroadcasterSessionDuration(seconds float64) {
	BroadcasterSessionDuration.Observe(seconds)
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.broadcasterSessionDuration.Record(context.Background(), seconds, metrics.options())
	}
}

func ObserveAPIRequestDuration(ctx context.Context, route, method, status string, seconds float64) {
	APIRequestDuration.WithLabelValues(route, method, status).Observe(seconds)
	if metrics := directMetrics.Load(); metrics != nil {
		metrics.apiRequestDuration.Record(ctx, seconds, metrics.options(
			attribute.String("route", route),
			attribute.String("method", method),
			attribute.String("status", status),
		))
	}
}
