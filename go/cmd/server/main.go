package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Ghost-15/streaming/internal/config"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/infrastructure/supabase"
	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
	"github.com/Ghost-15/streaming/internal/router"
	"github.com/Ghost-15/streaming/internal/usecase"
)

// main is the composition root: it wires all dependencies manually (no DI framework).
// ADR-002: explicit wiring > magic DI — easier to read and defend in soutenance.
func main() {
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	// 1. Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("config load failed")
	}

	// 2. Loki writer — multi-writer: stdout + Loki
	lokiWriter, lokiErr := telemetry.NewLokiWriter(
		os.Getenv("LOKI_URL"),
		os.Getenv("LOKI_USERNAME"),
		os.Getenv("LOKI_PASSWORD"),
		"streampulse-api",
		cfg.Env,
	)
	if lokiErr != nil {
		log.Warn().Err(lokiErr).Msg("loki unavailable, logging to stdout only")
	} else {
		defer lokiWriter.Close()
		multi := zerolog.MultiLevelWriter(os.Stdout, lokiWriter)
		log.Logger = zerolog.New(multi).With().Timestamp().Logger()
	}

	ctx := context.Background()

	// 3. OpenTelemetry — init provider (non-bloquant si collector indisponible)
	otelShutdown, err := telemetry.InitTracer(ctx, "streampulse-api", cfg.OTELServiceNamespace, cfg.OTELDeploymentEnv, cfg.OTELEndpoint)
	if err != nil {
		log.Warn().Err(err).Msg("otel unavailable, traces disabled")
	} else {
		defer func() {
			if err := otelShutdown(ctx); err != nil {
				log.Error().Err(err).Msg("otel shutdown failed")
			}
		}()
	}

	// 4. Infrastructure — database (non-bloquant si pas encore de BDD)
	db, err := supabase.NewPool(ctx, cfg.SupabaseDBURL)
	if err != nil {
		log.Warn().Err(err).Msg("database unavailable, api starts without db")
		db = nil
	}
	if db != nil {
		defer db.Close()
	}

	// 5. Repositories (infrastructure layer)
	userRepo := supabase.NewUserRepo(db)
	streamRepo := supabase.NewStreamRepo(db)
	playlistRepo := supabase.NewPlaylistRepo(db)

	// 6. Use Cases (business layer)
	authUC := usecase.NewAuthUseCase(userRepo, cfg.JWTPrivateKeyPath)
	streamUC := usecase.NewStreamUseCase(streamRepo)
	playlistUC := usecase.NewPlaylistUseCase(playlistRepo)

	// 7. Handlers (presentation layer)
	authH := handler.NewAuthHandler(authUC)
	streamH := handler.NewStreamHandler(streamUC)
	playlistH := handler.NewPlaylistHandler(playlistUC)

	// 8. Router
	engine := router.NewRouter(cfg, authH, streamH, playlistH)

	// 9. HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.Port).Str("env", cfg.Env).Msg("streampulse-api listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("http server failed")
		}
	}()

	// Graceful shutdown on SIGTERM / SIGINT (K8s sends SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}
	log.Info().Msg("server exited cleanly")
}
