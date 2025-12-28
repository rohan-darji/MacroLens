package main

import (
	"fmt"
	"log"
	"os"

	"github.com/macrolens/backend/config"
	httpDelivery "github.com/macrolens/backend/internal/delivery/http"
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

	// Create HTTP handler
	handler := httpDelivery.NewHandler()

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
