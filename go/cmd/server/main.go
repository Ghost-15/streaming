package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	// 1. Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	// 2. OpenTelemetry — init provider (non-bloquant si collector indisponible)
	otelShutdown, err := telemetry.InitTracer(ctx, "streampulse-api", cfg.OTELEndpoint)
	if err != nil {
		log.Printf("WARNING: OTEL unavailable (%v) — traces désactivées", err)
	} else {
		defer func() {
			if err := otelShutdown(ctx); err != nil {
				log.Printf("otel shutdown: %v", err)
			}
		}()
	}

	// 3. Infrastructure — database (non-bloquant si pas encore de BDD)
	db, err := supabase.NewPool(ctx, cfg.SupabaseDBURL)
	if err != nil {
		log.Printf("WARNING: database unavailable (%v) — API démarre sans BDD, les endpoints DB renverront 500", err)
		db = nil
	}
	if db != nil {
		defer db.Close()
	}

	// 4. Repositories (infrastructure layer)
	userRepo := supabase.NewUserRepo(db)
	streamRepo := supabase.NewStreamRepo(db)
	playlistRepo := supabase.NewPlaylistRepo(db)

	// 5. Use Cases (business layer)
	authUC := usecase.NewAuthUseCase(userRepo, cfg.JWTPrivateKeyPath)
	streamUC := usecase.NewStreamUseCase(streamRepo)
	playlistUC := usecase.NewPlaylistUseCase(playlistRepo)

	// 6. Handlers (presentation layer)
	authH := handler.NewAuthHandler(authUC)
	streamH := handler.NewStreamHandler(streamUC)
	playlistH := handler.NewPlaylistHandler(playlistUC)

	// 7. Router
	engine := router.NewRouter(cfg, authH, streamH, playlistH)

	// 8. HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("streampulse-api listening on :%s (env=%s)", cfg.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	// Graceful shutdown on SIGTERM / SIGINT (K8s sends SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server exited cleanly")
}
