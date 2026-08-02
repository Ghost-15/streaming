package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Ghost-15/streaming/internal/config"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/infrastructure/streaming"
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
	rootCtx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

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
	adminRepo := supabase.NewAdminRepo(db)
	favoriteRepo := supabase.NewFavoriteRepo(db)
	historyRepo := supabase.NewListenHistoryRepo(db)
	recommendationRepo := supabase.NewRecommendationRepo(db)

	// 6. Use Cases (business layer)
	authUC := usecase.NewAuthUseCase(userRepo, cfg.JWTPrivateKeyPath)
	streamUC := usecase.NewStreamUseCase(streamRepo, historyRepo)
	playlistUC := usecase.NewPlaylistUseCase(playlistRepo)
	adminUC := usecase.NewAdminUseCase(adminRepo)
	favoriteUC := usecase.NewFavoriteUseCase(favoriteRepo)
	recommendationUC := usecase.NewRecommendationUseCase(recommendationRepo)

	// 6b. Audio relay hub (goroutines + channels, no external dependency)
	hub := streaming.NewHub()

	// 7. Handlers (presentation layer)
	authH := handler.NewAuthHandler(authUC)
	audioHub := streaming.NewHub()
	streamH := handler.NewStreamHandler(
		streamUC,
		handler.WithAudioStreaming(
			audioHub,
			cfg.StreamMaxDuration,
			cfg.StreamIdleTimeout,
			cfg.StreamWriteTimeout,
			cfg.StreamMaxIngestBytes,
			cfg.StreamChunkSize,
			cfg.StreamClientBuffer,
		),
	)
	playlistH := handler.NewPlaylistHandler(playlistUC)
	adminH := handler.NewAdminHandler(adminUC)
	favoriteH := handler.NewFavoriteHandler(favoriteUC)
	recommendationH := handler.NewRecommendationHandler(recommendationUC)

	// 8. Router
	engine := router.NewRouter(cfg, authH, streamH, playlistH, adminH, favoriteH, recommendationH)

	// 9. HTTP server with graceful shutdown.
	// ReadTimeout and WriteTimeout are 0 (disabled) to allow long-lived
	// streaming connections (broadcaster Push + listener Audio).
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           engine,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		// ReadTimeout and WriteTimeout intentionally remain zero: absolute
		// server deadlines are incompatible with multi-hour audio requests.
		// The audio handlers apply sliding read/write deadlines instead.
		IdleTimeout: cfg.HTTPIdleTimeout,
		BaseContext: func(net.Listener) context.Context {
			return rootCtx
		},
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info().Str("port", cfg.Port).Str("env", cfg.Env).Msg("streampulse-api listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	var pprofServer *http.Server
	if cfg.PprofEnabled {
		pprofServer = telemetry.NewPprofServer(cfg.PprofAddr)
		go func() {
			log.Info().Str("address", cfg.PprofAddr).Msg("pprof diagnostics listening")
			if err := pprofServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error().Err(err).Msg("pprof server failed")
			}
		}()
	}

	select {
	case <-rootCtx.Done():
	case err := <-serverErr:
		log.Error().Err(err).Msg("http server failed")
		stopSignals()
	}

	log.Info().Msg("shutting down server")
	// Close the Hub first so long-lived responses finish before Server.Shutdown
	// waits for HTTP handlers.
	audioHub.Shutdown()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
		_ = srv.Close()
	}
	if pprofServer != nil {
		if err := pprofServer.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("pprof shutdown failed")
		}
	}
	log.Info().Msg("server exited cleanly")
}
