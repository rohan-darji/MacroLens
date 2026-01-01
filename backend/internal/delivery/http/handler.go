package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/macrolens/backend/internal/domain"
	"github.com/macrolens/backend/internal/usecase"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	nutritionService *usecase.NutritionService
}

// NewHandler creates a new HTTP handler with the given nutrition service.
// If nutritionService is nil, SearchNutrition will return 501 Not Implemented.
func NewHandler(nutritionService *usecase.NutritionService) *Handler {
	return &Handler{
		nutritionService: nutritionService,
	}
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
// POST /api/v1/nutrition/search
// Request body: { "productName": "...", "brand": "...", "size": "..." }
// Response: NutritionData or error
func (h *Handler) SearchNutrition(c *gin.Context) {
	// Check if service is available
	if h.nutritionService == nil {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "Nutrition search service not configured",
		})
		return
	}

	// Parse request body
	var request domain.SearchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if request.ProductName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "productName is required",
		})
		return
	}

	// Call nutrition service
	result, err := h.nutritionService.SearchNutrition(c.Request.Context(), &request)

	// Handle errors with appropriate HTTP status codes
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRequest):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, domain.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No matching product found in USDA database",
			})
		case errors.Is(err, domain.ErrLowConfidence):
			// Return data with warning for low confidence matches
			c.JSON(http.StatusOK, gin.H{
				"data":    result,
				"warning": "Low confidence match - verify the product manually",
			})
		case errors.Is(err, domain.ErrRateLimited):
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded, please try again later",
			})
		case errors.Is(err, domain.ErrUSDAAPIFailure):
			c.JSON(http.StatusBadGateway, gin.H{
				"error": "USDA API temporarily unavailable",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "An unexpected error occurred",
			})
		}
		return
	}

	// Success - return nutrition data
	c.JSON(http.StatusOK, result)
}
