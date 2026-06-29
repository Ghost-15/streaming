package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler/middleware"
	"github.com/Ghost-15/streaming/internal/usecase"
)

// AdminHandler handles HTTP requests for admin operations.
type AdminHandler struct {
	useCase usecase.AdminUseCase
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(uc usecase.AdminUseCase) *AdminHandler {
	return &AdminHandler{useCase: uc}
}

// UpdateRoleRequest is the JSON body for PUT /admin/users/:id/role.
type UpdateRoleRequest struct {
	Role entity.UserRole `json:"role" binding:"required"`
}

// ListUsers godoc
// @Summary     List all users (paginated)
// @Tags        admin
// @Produce     json
// @Param       page  query int false "Page number" default(1)
// @Param       limit query int false "Items per page" default(20)
// @Success     200 {object} map[string]interface{}
// @Router      /api/v1/admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	users, total, err := h.useCase.ListUsers(c.Request.Context(), page, limit)
	if err != nil {
		middleware.Logger(c).Error().Err(err).Msg("admin: list users failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	middleware.Logger(c).Info().Int("total", total).Int("page", page).Msg("admin: listed users")
	c.JSON(http.StatusOK, gin.H{
		"data":  users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetUser godoc
// @Summary     Get user by ID
// @Tags        admin
// @Produce     json
// @Param       id path string true "User ID"
// @Success     200 {object} entity.User
// @Router      /api/v1/admin/users/{id} [get]
func (h *AdminHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.useCase.GetUser(c.Request.Context(), id)
	if err != nil {
		middleware.Logger(c).Error().Err(err).Str("user_id", id).Msg("admin: get user failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUserRole godoc
// @Summary     Update user role
// @Tags        admin
// @Accept      json
// @Produce     json
// @Param       id   path string           true "User ID"
// @Param       body body UpdateRoleRequest true "New role"
// @Success     200 {object} map[string]string
// @Router      /api/v1/admin/users/{id}/role [put]
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	id := c.Param("id")

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.UpdateUserRole(c.Request.Context(), id, req.Role); err != nil {
		middleware.Logger(c).Error().Err(err).Str("user_id", id).Msg("admin: update role failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Audit log
	claims, _ := middleware.GetClaims(c)
	if claims != nil {
		middleware.Logger(c).Info().
			Str("admin_id", claims.UserID).
			Str("target_user_id", id).
			Str("new_role", string(req.Role)).
			Msg("admin: role updated")
	}

	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

// SuspendUser godoc
// @Summary     Suspend or reactivate a user account
// @Tags        admin
// @Produce     json
// @Param       id      path string true "User ID"
// @Param       suspend query bool   true "true to suspend, false to reactivate"
// @Success     200 {object} map[string]string
// @Router      /api/v1/admin/users/{id}/suspend [post]
func (h *AdminHandler) SuspendUser(c *gin.Context) {
	id := c.Param("id")
	suspend := c.Query("suspend") != "false"

	if err := h.useCase.SuspendUser(c.Request.Context(), id, suspend); err != nil {
		middleware.Logger(c).Error().Err(err).Str("user_id", id).Msg("admin: suspend user failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Audit log
	claims, _ := middleware.GetClaims(c)
	if claims != nil {
		action := "suspended"
		if !suspend {
			action = "reactivated"
		}
		middleware.Logger(c).Info().
			Str("admin_id", claims.UserID).
			Str("target_user_id", id).
			Str("action", action).
			Msg("admin: user account updated")
	}

	c.JSON(http.StatusOK, gin.H{"message": "user account updated"})
}

// GetStats godoc
// @Summary     Get admin statistics
// @Tags        admin
// @Produce     json
// @Success     200 {object} entity.AdminStats
// @Router      /api/v1/admin/stats [get]
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.useCase.GetStats(c.Request.Context())
	if err != nil {
		middleware.Logger(c).Error().Err(err).Msg("admin: get stats failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	middleware.Logger(c).Info().Int("total_users", stats.TotalUsers).Msg("admin: stats fetched")
	c.JSON(http.StatusOK, stats)
}
