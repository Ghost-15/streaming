package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/handler/middleware"
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
	Email     string `json:"email"      binding:"required,email"`
	Password  string `json:"password"   binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// LoginRequest is the JSON body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// RefreshRequest is the JSON body for POST /auth/refresh and POST /auth/logout.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// sessionPayload renders the credentials of an authenticated session.
// The access token keeps the historical "token" key so existing clients keep working.
func sessionPayload(result *usecase.AuthResult) gin.H {
	return gin.H{
		"token":         result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_in":    result.ExpiresIn,
		"user":          result.User,
	}
}

// Register godoc.
// @Summary     Register a new user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body RegisterRequest true "Register payload"
// @Success     201 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("invalid register payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.useCase.Register(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName, req.Role)
	if err != nil {
		middleware.Logger(c).Error().Err(err).Msg("register failed")
		// Sentinel error check for email already registered
		if errors.Is(err, usecase.ErrEmailAlreadyInUse) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		// Other errors → 500
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.Logger(c).Info().Str("user_id", result.User.ID).Msg("user registered")

	c.JSON(http.StatusCreated, sessionPayload(result))
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

	result, err := h.useCase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("login rejected")
		if errors.Is(err, usecase.ErrAccountSuspended) {
			c.JSON(http.StatusForbidden, gin.H{"error": "account suspended"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	middleware.Logger(c).Info().Msg("user authenticated")

	c.JSON(http.StatusOK, sessionPayload(result))
}

// Refresh godoc.
// @Summary     Exchange a refresh token for a new token pair
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body RefreshRequest true "Refresh payload"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Router      /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("invalid refresh payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.useCase.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("refresh rejected")
		if errors.Is(err, usecase.ErrAccountSuspended) {
			c.JSON(http.StatusForbidden, gin.H{"error": "account suspended"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	middleware.Logger(c).Info().Str("user_id", result.User.ID).Msg("session refreshed")

	c.JSON(http.StatusOK, sessionPayload(result))
}

// Logout godoc.
// @Summary     Revoke a refresh token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body RefreshRequest true "Logout payload"
// @Success     204
// @Failure     400 {object} map[string]string
// @Router      /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("invalid logout payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		middleware.Logger(c).Error().Err(err).Msg("logout failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.Logger(c).Info().Msg("refresh token revoked")

	c.Status(http.StatusNoContent)
}

// Me godoc.
// @Summary     Return the authenticated user
// @Tags        auth
// @Produce     json
// @Success     200 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	uid, ok := ownerID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	user, err := h.useCase.Me(c.Request.Context(), uid)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		middleware.Logger(c).Error().Err(err).Msg("read profile failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// ChangePasswordRequest is the JSON body for PUT /auth/password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=8"`
	NewPassword     string `json:"new_password"     binding:"required,min=8"`
}

// ChangePassword godoc.
// @Summary     Change the password of the authenticated user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body ChangePasswordRequest true "Password payload"
// @Success     204
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Router      /api/v1/auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	uid, ok := ownerID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("invalid password payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.useCase.ChangePassword(c.Request.Context(), uid, req.CurrentPassword, req.NewPassword)
	switch {
	case err == nil:
	case errors.Is(err, usecase.ErrInvalidCredentials):
		middleware.Logger(c).Warn().Msg("password change rejected")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	case errors.Is(err, usecase.ErrPasswordUnchanged):
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must differ from the current one"})
		return
	case errors.Is(err, usecase.ErrPasswordTooShort):
		c.JSON(http.StatusBadRequest, gin.H{"error": "password too short"})
		return
	case errors.Is(err, usecase.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	default:
		middleware.Logger(c).Error().Err(err).Msg("password change failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.Logger(c).Info().Str("user_id", uid).Msg("password changed")

	c.Status(http.StatusNoContent)
}

// DeleteMe godoc.
// @Summary     Delete the authenticated account (RGPD right to erasure)
// @Tags        auth
// @Produce     json
// @Success     204
// @Failure     401 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /api/v1/auth/me [delete]
func (h *AuthHandler) DeleteMe(c *gin.Context) {
	uid, ok := ownerID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	if err := h.useCase.DeleteAccount(c.Request.Context(), uid); err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		middleware.Logger(c).Error().Err(err).Msg("delete account failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.Logger(c).Info().Str("user_id", uid).Msg("account deleted")

	c.Status(http.StatusNoContent)
}
