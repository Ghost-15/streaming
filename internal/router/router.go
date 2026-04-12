package router

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/handler/middleware"
)

// NewRouter builds the Gin engine with all routes and middlewares.
// Called once in main.go after dependency injection.
// ADR-001: gin.New() (not gin.Default()) — middlewares are explicit.
func NewRouter(
	authH *handler.AuthHandler,
	streamH *handler.StreamHandler,
	playlistH *handler.PlaylistHandler,
) *gin.Engine {
	r := gin.New()

	// Global middlewares (order matters)
	r.Use(otelgin.Middleware("streampulse-api")) // Sprint 2 — US-008
	r.Use(middleware.ZerologMiddleware())         // Sprint 1 — US-006
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Recovery())

	// Public routes — no auth required
	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		auth.Use(middleware.RateLimitMiddleware(5, 3)) // 5 req/min, burst 3
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
		}
	}

	// Protected routes — JWT + RBAC required
	protected := r.Group("/api/v1")
	protected.Use(middleware.RBACMiddleware(
		entity.RoleUser,
		entity.RoleDiffuseur,
		entity.RoleAdmin,
	))
	{
		protected.GET("/streams", streamH.ListActive)
		protected.GET("/streams/:id/listen", streamH.Listen)

		diffuseur := protected.Group("/")
		diffuseur.Use(middleware.RBACMiddleware(entity.RoleDiffuseur, entity.RoleAdmin))
		{
			diffuseur.POST("/streams", streamH.Start)
		}

		protected.GET("/playlists", playlistH.List)
		protected.POST("/playlists", playlistH.Create)
		protected.DELETE("/playlists/:id", playlistH.Delete)
	}

	// Admin routes
	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.RBACMiddleware(entity.RoleAdmin))
	admin.Use(middleware.RateLimitMiddleware(20, 10))
	{
		// TODO Sprint 3 — US-013: admin panel routes
	}

	// Health check (no auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "streampulse-api"})
	})

	return r
}
