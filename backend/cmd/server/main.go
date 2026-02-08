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

	// Enable debug mode in development environment
	if cfg.Server.Environment == "development" {
		usdaClient.SetDebug(true)
		log.Printf("USDA client debug mode enabled")
	}

	if cfg.USDA.APIKey != "" {
		log.Printf("USDA API configured: %s (key: %s...)", cfg.USDA.BaseURL, cfg.USDA.APIKey[:8])
	} else {
		log.Printf("WARNING: USDA API configured: %s (key: NOT CONFIGURED - API calls will fail!)", cfg.USDA.BaseURL)
	}

	// Initialize usecase layer
	nutritionService := usecase.NewNutritionService(
		memoryCache,
		usdaClient,
		usecase.NutritionServiceConfig{
			CacheTTL:               cfg.Cache.TTL,
			MinConfidenceThreshold: cfg.Matching.MinConfidenceThreshold,
			EnableFuzzyMatching:    cfg.Matching.EnableFuzzyMatching,
			EnableDebugLogging:     cfg.Matching.EnableDebugLogging,
		},
	)

	log.Printf("Matching: confidence=%.0f%%, fuzzy=%v, debug=%v",
		cfg.Matching.MinConfidenceThreshold,
		cfg.Matching.EnableFuzzyMatching,
		cfg.Matching.EnableDebugLogging)

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
