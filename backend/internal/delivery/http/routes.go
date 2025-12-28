package http

import (
	"github.com/gin-gonic/gin"
	"github.com/macrolens/backend/config"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(cfg *config.Config, handler *Handler) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(RecoveryMiddleware())
	router.Use(LoggerMiddleware())
	router.Use(CORSMiddleware(cfg.Server.AllowedOrigins))

	// Health check endpoint
	router.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Nutrition endpoints
		nutrition := v1.Group("/nutrition")
		{
			nutrition.POST("/search", handler.SearchNutrition)
			// TODO: Add more endpoints in Phase 2
			// nutrition.GET("/:fdcId", handler.GetNutritionByID)
		}
	}

	return router
}
