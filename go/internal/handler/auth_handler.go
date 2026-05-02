package handler

import (
	"net/http"

	"github.com/Ghost-15/streaming/internal/handler/middleware"
	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/usecase"
)

// AuthHandler handles HTTP requests for authentication.
// Sprint 1 — US-001.
type AuthHandler struct {
	useCase usecase.AuthUseCase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(uc usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{useCase: uc}
}

// RegisterRequest is the JSON body for POST /auth/register.
type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest is the JSON body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// Register godoc.
// @Summary     Register a new user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body RegisterRequest true "Register payload"
// @Success     201 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Router      /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("invalid register payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.useCase.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		middleware.Logger(c).Error().Err(err).Msg("register failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.Logger(c).Info().Str("user_id", user.ID).Msg("user registered")

	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "email": user.Email})
}

// Login godoc.
// @Summary     Authenticate and receive a JWT token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body LoginRequest true "Login payload"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Router      /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("invalid login payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.useCase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("login rejected")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	middleware.Logger(c).Info().Msg("user authenticated")

	c.JSON(http.StatusOK, gin.H{"token": token})
}
