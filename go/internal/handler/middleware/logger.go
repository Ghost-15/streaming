package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

// ZerologMiddleware logs each request as a structured JSON line.
// It also injects a request-scoped zerolog logger into the request context.
// Logs are forwarded to Loki via the global zerolog.Logger MultiLevelWriter
// configured in main.go — no extra setup needed here.
func ZerologMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		traceID := traceIDFromContext(c.Request.Context())
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		// Enrich the request-scoped logger with fields useful for Loki queries:
		// {service="streampulse-api", route="/api/v1/auth/login"}
		requestLogger := log.Logger.With().
			Str("trace_id", traceID).
			Str("route", route).
			Str("method", c.Request.Method).
			Logger()

		c.Request = c.Request.WithContext(requestLogger.WithContext(c.Request.Context()))

		c.Next()

		userID := ""
		if claims, ok := GetClaims(c); ok {
			userID = claims.UserID
		}

		status := c.Writer.Status()
		fields := requestLogger.With().
			Str("user_id", userID).
			Int("status", status).
			Dur("duration", time.Since(start)).
			Logger()

		switch {
		case status >= 500 || len(c.Errors) > 0:
			fields.Error().Msg("request")
		case status >= 400:
			fields.Warn().Msg("request")
		default:
			fields.Info().Msg("request")
		}
	}
}

// Logger returns the request-scoped zerolog logger stored in the request context.
// This logger inherits the global MultiLevelWriter (stdout + Loki).
func Logger(c *gin.Context) *zerolog.Logger {
	if logger := zerolog.Ctx(c.Request.Context()); logger != nil {
		return logger
	}
	return &log.Logger
}

func traceIDFromContext(ctx context.Context) string {
	spanContext := trace.SpanFromContext(ctx).SpanContext()
	if !spanContext.IsValid() {
		return "no-trace"
	}
	return spanContext.TraceID().String()
}
