package router

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	_ "github.com/Ghost-15/streaming/docs/openapi" // generated OpenAPI spec served by /swagger
	"github.com/Ghost-15/streaming/internal/config"
	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/handler/middleware"
)

// NewRouter builds the Gin engine with all routes and middlewares.
func NewRouter(
	cfg *config.Config,
	authH *handler.AuthHandler,
	streamH *handler.StreamHandler,
	playlistH *handler.PlaylistHandler,
	adminH *handler.AdminHandler,
	favoriteH *handler.FavoriteHandler,
	recommendationH *handler.RecommendationHandler,
) *gin.Engine {
	pubKeyBytes, err := os.ReadFile(cfg.JWTPublicKeyPath)
	if err != nil {
		panic(fmt.Sprintf("router: read public key %q: %v", cfg.JWTPublicKeyPath, err))
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		panic(fmt.Sprintf("router: parse public key: %v", err))
	}

	r := gin.New()

	r.Use(otelgin.Middleware("streampulse-api"))
	r.Use(middleware.ZerologMiddleware())
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.MetricsMiddleware())
	r.Use(middleware.CORSMiddleware(cfg.CORSOrigins))
	r.Use(gin.Recovery())

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		auth.Use(middleware.RateLimitMiddleware(5, 5))
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
			auth.POST("/refresh", authH.Refresh)
			auth.POST("/logout", authH.Logout)
		}

		v1.GET("/streams", streamH.ListActive)
		// Public: browser <audio> cannot set Authorization headers
		v1.GET("/streams/:id/audio", streamH.Audio)
		v1.GET("/streams/:id/audio/ws", streamH.AudioSocket)
	}

	protected := r.Group("/api/v1")
	protected.Use(middleware.RBACMiddleware(publicKey,
		entity.RoleUser,
		entity.RoleDiffuseur,
		entity.RoleAdmin,
	))
	protected.Use(middleware.UserRateLimitMiddleware(100, 100))
	{
		protected.GET("/auth/me", authH.Me)
		protected.DELETE("/auth/me", authH.DeleteMe)
		protected.PUT("/auth/password", authH.ChangePassword)

		protected.GET("/streams/:id/listen", streamH.StreamAudio)
		protected.POST("/streams/:id/listen", streamH.Listen)
		protected.POST("/streams/:id/leave", streamH.Leave)

		diffuseur := protected.Group("/")
		diffuseur.Use(middleware.RBACMiddleware(publicKey, entity.RoleDiffuseur, entity.RoleAdmin))
		{
			diffuseur.GET("/streams/mine", streamH.ListOwned)
			diffuseur.POST("/streams", streamH.Start)
			diffuseur.PUT("/streams/:id/start", streamH.Restart)
			diffuseur.PUT("/streams/:id/stop", streamH.Stop)
			diffuseur.DELETE("/streams/:id", streamH.Delete)
		}

		protected.GET("/playlists", playlistH.List)
		protected.POST("/playlists", playlistH.Create)
		protected.GET("/playlists/:id", playlistH.GetByID)
		protected.PUT("/playlists/:id", playlistH.Update)
		protected.DELETE("/playlists/:id", playlistH.Delete)
		protected.POST("/playlists/:id/tracks", playlistH.AddTrack)
		protected.DELETE("/playlists/:id/tracks/:trackID", playlistH.RemoveTrack)
		protected.PUT("/playlists/:id/tracks/reorder", playlistH.ReorderTracks)
		protected.POST("/playlists/:id/next", playlistH.NextTrack)

		protected.GET("/favorites", favoriteH.List)
		protected.POST("/favorites", favoriteH.Add)
		protected.DELETE("/favorites/:streamID", favoriteH.Remove)

		protected.GET("/recommendations", recommendationH.List)
	}

	// MediaRecorder sends two short audio requests per second. Keep this data
	// plane outside the 100 req/min business-API limiter, otherwise its token
	// bucket is eventually exhausted and a healthy live is stopped by a 429.
	media := r.Group("/api/v1")
	media.Use(middleware.RBACMiddleware(publicKey, entity.RoleDiffuseur, entity.RoleAdmin))
	media.Use(middleware.StreamDataRateLimitMiddleware())
	{
		media.POST("/streams/:id/push", streamH.PushAudio)
		media.PUT("/streams/:id/audio", streamH.IngestAudio)
	}

	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.RBACMiddleware(publicKey, entity.RoleAdmin))
	admin.Use(middleware.UserRateLimitMiddleware(100, 100))
	admin.Use(middleware.RateLimitMiddleware(20, 10))
	{
		admin.GET("/users", adminH.ListUsers)
		admin.GET("/users/:id", adminH.GetUser)
		admin.PUT("/users/:id/role", adminH.UpdateUserRole)
		admin.POST("/users/:id/suspend", adminH.SuspendUser)
		admin.DELETE("/users/:id", adminH.DeleteUser)
		admin.GET("/stats", adminH.GetStats)
	}

	r.GET("/health", handler.Health)

	// Interactive OpenAPI contract. Documentation only: each route it lists is
	// still guarded by its own middleware.
	if cfg.SwaggerEnabled {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Prometheus metrics endpoint (Sprint 3 — US-010, scraped by Prometheus).
	r.GET("/metrics", middleware.MetricsAuthMiddleware(cfg.MetricsBearerToken), gin.WrapH(promhttp.Handler()))

	return r
}
