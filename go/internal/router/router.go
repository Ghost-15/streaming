package router

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

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
	r.Use(middleware.CORSMiddleware(cfg.CORSOrigins))
	r.Use(gin.Recovery())

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		auth.Use(middleware.RateLimitMiddleware(5, 5))
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
		}
	}

	protected := r.Group("/api/v1")
	protected.Use(middleware.RBACMiddleware(publicKey,
		entity.RoleUser,
		entity.RoleDiffuseur,
		entity.RoleAdmin,
	))
	protected.Use(middleware.UserRateLimitMiddleware(100, 100))
	{
		protected.GET("/streams", streamH.ListActive)
		protected.GET("/streams/:id/listen", streamH.Listen)

		diffuseur := protected.Group("/")
		diffuseur.Use(middleware.RBACMiddleware(publicKey, entity.RoleDiffuseur, entity.RoleAdmin))
		{
			diffuseur.POST("/streams", streamH.Start)
		}

		protected.GET("/playlists", playlistH.List)
		protected.POST("/playlists", playlistH.Create)
		protected.GET("/playlists/:id", playlistH.GetByID)
		protected.PUT("/playlists/:id", playlistH.Update)
		protected.DELETE("/playlists/:id", playlistH.Delete)
		protected.POST("/playlists/:id/tracks", playlistH.AddTrack)
		protected.DELETE("/playlists/:id/tracks/:trackID", playlistH.RemoveTrack)
	}

	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.RBACMiddleware(publicKey, entity.RoleAdmin))
	admin.Use(middleware.UserRateLimitMiddleware(100, 100))
	admin.Use(middleware.RateLimitMiddleware(20, 10))
	{
		// TODO Sprint 3: admin panel routes
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "streampulse-api"})
	})

	return r
}
