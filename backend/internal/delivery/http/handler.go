package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	// TODO: Add nutrition usecase when implemented
}

// NewHandler creates a new HTTP handler
func NewHandler() *Handler {
	return &Handler{}
}

// HealthCheck returns the health status of the API
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "macrolens-backend",
		"version": "1.0.0",
	})
}

// SearchNutrition handles nutrition search requests
// TODO: Implement this in Phase 2
func (h *Handler) SearchNutrition(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Nutrition search not yet implemented - coming in Phase 2",
	})
}
