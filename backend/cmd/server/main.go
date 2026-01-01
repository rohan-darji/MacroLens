package main

import (
	"fmt"
	"log"
	"os"

	"github.com/macrolens/backend/config"
	httpDelivery "github.com/macrolens/backend/internal/delivery/http"
	"github.com/macrolens/backend/internal/infrastructure/cache"
	"github.com/macrolens/backend/internal/infrastructure/usda"
	"github.com/macrolens/backend/internal/usecase"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting MacroLens Backend v1.0.0")
	log.Printf("Environment: %s", cfg.Server.Environment)
	log.Printf("Port: %s", cfg.Server.Port)
	log.Printf("Cache Type: %s", cfg.Cache.Type)

	// Initialize infrastructure dependencies
	memoryCache := cache.NewMemoryCache()
	log.Printf("Cache TTL: %s", cfg.Cache.TTL)

	usdaClient := usda.NewClient(cfg.USDA.APIKey, cfg.USDA.BaseURL)
	if cfg.USDA.APIKey != "" {
		log.Printf("USDA API configured: %s (key: configured)", cfg.USDA.BaseURL)
	} else {
		log.Printf("USDA API configured: %s (key: not configured)", cfg.USDA.BaseURL)
	}

	// Initialize usecase layer
	nutritionService := usecase.NewNutritionService(
		memoryCache,
		usdaClient,
		usecase.NutritionServiceConfig{
			CacheTTL:               cfg.Cache.TTL,
			MinConfidenceThreshold: 40, // 40% minimum match confidence
		},
	)

	// Create HTTP handler with dependencies
	handler := httpDelivery.NewHandler(nutritionService)

	// Setup router
	router := httpDelivery.SetupRouter(cfg, handler)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server listening on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func init() {
	// Set log flags for better debugging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)
}
