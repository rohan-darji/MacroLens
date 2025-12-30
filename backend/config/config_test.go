package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Clean up environment before tests
	cleanupEnv := func() {
		os.Unsetenv("MACROLENS_SERVER_PORT")
		os.Unsetenv("MACROLENS_SERVER_ENVIRONMENT")
		os.Unsetenv("MACROLENS_SERVER_ALLOWED_ORIGINS")
		os.Unsetenv("MACROLENS_USDA_API_KEY")
		os.Unsetenv("MACROLENS_USDA_BASE_URL")
		os.Unsetenv("MACROLENS_CACHE_TYPE")
		os.Unsetenv("MACROLENS_CACHE_REDIS_URL")
		os.Unsetenv("MACROLENS_CACHE_TTL")
		os.Unsetenv("MACROLENS_RATELIMIT_PER_IP")
		os.Unsetenv("MACROLENS_RATELIMIT_USDA")
	}

	t.Run("loads with defaults when no env vars set", func(t *testing.T) {
		cleanupEnv()
		// Set required API key
		os.Setenv("MACROLENS_USDA_API_KEY", "test-key")
		defer cleanupEnv()

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}

		// Check defaults
		if cfg.Server.Port != "8080" {
			t.Errorf("Server.Port = %s, want 8080", cfg.Server.Port)
		}
		if cfg.Server.Environment != "development" {
			t.Errorf("Server.Environment = %s, want development", cfg.Server.Environment)
		}
		if cfg.USDA.BaseURL != "https://api.nal.usda.gov/fdc" {
			t.Errorf("USDA.BaseURL = %s, want https://api.nal.usda.gov/fdc", cfg.USDA.BaseURL)
		}
		if cfg.Cache.Type != "memory" {
			t.Errorf("Cache.Type = %s, want memory", cfg.Cache.Type)
		}
		if cfg.Cache.TTL != 720*time.Hour {
			t.Errorf("Cache.TTL = %v, want 720h", cfg.Cache.TTL)
		}
		if cfg.RateLimit.PerIP != 100 {
			t.Errorf("RateLimit.PerIP = %d, want 100", cfg.RateLimit.PerIP)
		}
		if cfg.RateLimit.USDA != 1000 {
			t.Errorf("RateLimit.USDA = %d, want 1000", cfg.RateLimit.USDA)
		}
	})

	t.Run("loads custom values from environment variables", func(t *testing.T) {
		cleanupEnv()
		os.Setenv("MACROLENS_SERVER_PORT", "9090")
		os.Setenv("MACROLENS_SERVER_ENVIRONMENT", "production")
		os.Setenv("MACROLENS_USDA_API_KEY", "custom-api-key")
		os.Setenv("MACROLENS_USDA_BASE_URL", "https://custom.api.com")
		os.Setenv("MACROLENS_CACHE_TYPE", "redis")
		os.Setenv("MACROLENS_CACHE_REDIS_URL", "redis://localhost:6379")
		os.Setenv("MACROLENS_CACHE_TTL", "24h")
		os.Setenv("MACROLENS_RATELIMIT_PER_IP", "200")
		os.Setenv("MACROLENS_RATELIMIT_USDA", "2000")
		defer cleanupEnv()

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}

		if cfg.Server.Port != "9090" {
			t.Errorf("Server.Port = %s, want 9090", cfg.Server.Port)
		}
		if cfg.Server.Environment != "production" {
			t.Errorf("Server.Environment = %s, want production", cfg.Server.Environment)
		}
		if cfg.USDA.APIKey != "custom-api-key" {
			t.Errorf("USDA.APIKey = %s, want custom-api-key", cfg.USDA.APIKey)
		}
		if cfg.USDA.BaseURL != "https://custom.api.com" {
			t.Errorf("USDA.BaseURL = %s, want https://custom.api.com", cfg.USDA.BaseURL)
		}
		if cfg.Cache.Type != "redis" {
			t.Errorf("Cache.Type = %s, want redis", cfg.Cache.Type)
		}
		if cfg.Cache.RedisURL != "redis://localhost:6379" {
			t.Errorf("Cache.RedisURL = %s, want redis://localhost:6379", cfg.Cache.RedisURL)
		}
		if cfg.Cache.TTL != 24*time.Hour {
			t.Errorf("Cache.TTL = %v, want 24h", cfg.Cache.TTL)
		}
		if cfg.RateLimit.PerIP != 200 {
			t.Errorf("RateLimit.PerIP = %d, want 200", cfg.RateLimit.PerIP)
		}
		if cfg.RateLimit.USDA != 2000 {
			t.Errorf("RateLimit.USDA = %d, want 2000", cfg.RateLimit.USDA)
		}
	})

	t.Run("fails validation when API key is missing", func(t *testing.T) {
		cleanupEnv()
		defer cleanupEnv()

		_, err := Load()
		if err == nil {
			t.Error("Load() error = nil, want error for missing API key")
		}
		if err != nil && err.Error() != "invalid configuration: USDA API key is required (set MACROLENS_USDA_API_KEY)" {
			t.Errorf("Load() error = %v, want 'USDA API key is required'", err)
		}
	})

	t.Run("fails validation for invalid cache type", func(t *testing.T) {
		cleanupEnv()
		os.Setenv("MACROLENS_USDA_API_KEY", "test-key")
		os.Setenv("MACROLENS_CACHE_TYPE", "invalid")
		defer cleanupEnv()

		_, err := Load()
		if err == nil {
			t.Error("Load() error = nil, want error for invalid cache type")
		}
	})

	t.Run("fails validation when redis URL missing for redis cache", func(t *testing.T) {
		cleanupEnv()
		os.Setenv("MACROLENS_USDA_API_KEY", "test-key")
		os.Setenv("MACROLENS_CACHE_TYPE", "redis")
		defer cleanupEnv()

		_, err := Load()
		if err == nil {
			t.Error("Load() error = nil, want error for missing Redis URL")
		}
	})
}

func TestLoadEnvFile(t *testing.T) {
	t.Run("returns nil when .env file doesn't exist", func(t *testing.T) {
		// Save current directory
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create temp directory
		tempDir := t.TempDir()
		os.Chdir(tempDir)

		err := loadEnvFile()
		if err != nil {
			t.Errorf("loadEnvFile() error = %v, want nil when file doesn't exist", err)
		}
	})

	t.Run("loads variables from .env file", func(t *testing.T) {
		// Save current directory
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create temp directory
		tempDir := t.TempDir()
		os.Chdir(tempDir)

		// Create .env file
		envContent := `
# Comment line
TEST_VAR_1=value1
TEST_VAR_2=value2

# Another comment
TEST_VAR_3=value3
`
		err := os.WriteFile(".env", []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}

		// Clear any existing values
		os.Unsetenv("TEST_VAR_1")
		os.Unsetenv("TEST_VAR_2")
		os.Unsetenv("TEST_VAR_3")

		err = loadEnvFile()
		if err != nil {
			t.Fatalf("loadEnvFile() error = %v, want nil", err)
		}

		if os.Getenv("TEST_VAR_1") != "value1" {
			t.Errorf("TEST_VAR_1 = %s, want value1", os.Getenv("TEST_VAR_1"))
		}
		if os.Getenv("TEST_VAR_2") != "value2" {
			t.Errorf("TEST_VAR_2 = %s, want value2", os.Getenv("TEST_VAR_2"))
		}
		if os.Getenv("TEST_VAR_3") != "value3" {
			t.Errorf("TEST_VAR_3 = %s, want value3", os.Getenv("TEST_VAR_3"))
		}

		// Cleanup
		os.Unsetenv("TEST_VAR_1")
		os.Unsetenv("TEST_VAR_2")
		os.Unsetenv("TEST_VAR_3")
	})

	t.Run("skips empty lines and comments", func(t *testing.T) {
		// Save current directory
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create temp directory
		tempDir := t.TempDir()
		os.Chdir(tempDir)

		// Create .env file with various formats
		envContent := `
# This is a comment
   # This is also a comment

TEST_SKIP_1=value1

TEST_SKIP_2=value2
# TEST_COMMENTED=should_not_load
`
		err := os.WriteFile(".env", []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}

		os.Unsetenv("TEST_SKIP_1")
		os.Unsetenv("TEST_SKIP_2")
		os.Unsetenv("TEST_COMMENTED")

		err = loadEnvFile()
		if err != nil {
			t.Fatalf("loadEnvFile() error = %v, want nil", err)
		}

		if os.Getenv("TEST_SKIP_1") != "value1" {
			t.Errorf("TEST_SKIP_1 not loaded correctly")
		}
		if os.Getenv("TEST_SKIP_2") != "value2" {
			t.Errorf("TEST_SKIP_2 not loaded correctly")
		}
		if os.Getenv("TEST_COMMENTED") != "" {
			t.Errorf("TEST_COMMENTED should not be loaded from comment")
		}

		os.Unsetenv("TEST_SKIP_1")
		os.Unsetenv("TEST_SKIP_2")
	})

	t.Run("doesn't override existing environment variables", func(t *testing.T) {
		// Save current directory
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create temp directory
		tempDir := t.TempDir()
		os.Chdir(tempDir)

		// Set existing env var
		os.Setenv("TEST_OVERRIDE", "existing-value")

		// Create .env file that tries to override
		envContent := "TEST_OVERRIDE=new-value"
		err := os.WriteFile(".env", []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}

		err = loadEnvFile()
		if err != nil {
			t.Fatalf("loadEnvFile() error = %v, want nil", err)
		}

		// Should still have original value
		if os.Getenv("TEST_OVERRIDE") != "existing-value" {
			t.Errorf("TEST_OVERRIDE = %s, want existing-value (should not override)", os.Getenv("TEST_OVERRIDE"))
		}

		os.Unsetenv("TEST_OVERRIDE")
	})
}

func TestValidate(t *testing.T) {
	t.Run("validates successfully with all required fields", func(t *testing.T) {
		cfg := &Config{
			USDA: USDAConfig{
				APIKey:  "test-key",
				BaseURL: "https://api.nal.usda.gov/fdc",
			},
			Cache: CacheConfig{
				Type: "memory",
			},
		}

		err := validate(cfg)
		if err != nil {
			t.Errorf("validate() error = %v, want nil", err)
		}
	})

	t.Run("fails when API key is empty", func(t *testing.T) {
		cfg := &Config{
			USDA: USDAConfig{
				APIKey: "",
			},
			Cache: CacheConfig{
				Type: "memory",
			},
		}

		err := validate(cfg)
		if err == nil {
			t.Error("validate() error = nil, want error for empty API key")
		}
	})

	t.Run("fails for invalid cache type", func(t *testing.T) {
		cfg := &Config{
			USDA: USDAConfig{
				APIKey: "test-key",
			},
			Cache: CacheConfig{
				Type: "invalid-type",
			},
		}

		err := validate(cfg)
		if err == nil {
			t.Error("validate() error = nil, want error for invalid cache type")
		}
	})

	t.Run("validates redis cache type with URL", func(t *testing.T) {
		cfg := &Config{
			USDA: USDAConfig{
				APIKey: "test-key",
			},
			Cache: CacheConfig{
				Type:     "redis",
				RedisURL: "redis://localhost:6379",
			},
		}

		err := validate(cfg)
		if err != nil {
			t.Errorf("validate() error = %v, want nil for valid redis config", err)
		}
	})

	t.Run("fails for redis cache without URL", func(t *testing.T) {
		cfg := &Config{
			USDA: USDAConfig{
				APIKey: "test-key",
			},
			Cache: CacheConfig{
				Type:     "redis",
				RedisURL: "",
			},
		}

		err := validate(cfg)
		if err == nil {
			t.Error("validate() error = nil, want error for redis without URL")
		}
	})
}
