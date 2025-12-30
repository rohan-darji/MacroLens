package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

// cleanupConfigEnv removes all MACROLENS_* environment variables
func cleanupConfigEnv(t *testing.T) {
	t.Helper()
	envVars := []string{
		"MACROLENS_SERVER_PORT",
		"MACROLENS_SERVER_ENVIRONMENT",
		"MACROLENS_SERVER_ALLOWED_ORIGINS",
		"MACROLENS_USDA_API_KEY",
		"MACROLENS_USDA_BASE_URL",
		"MACROLENS_CACHE_TYPE",
		"MACROLENS_CACHE_REDIS_URL",
		"MACROLENS_CACHE_TTL",
		"MACROLENS_RATELIMIT_PER_IP",
		"MACROLENS_RATELIMIT_USDA",
	}
	for _, env := range envVars {
		os.Unsetenv(env)
	}
}

func TestLoad(t *testing.T) {

	t.Run("loads with defaults when no env vars set", func(t *testing.T) {
		cleanupConfigEnv(t)
		t.Cleanup(func() { cleanupConfigEnv(t) })

		// Set required API key
		os.Setenv("MACROLENS_USDA_API_KEY", "test-key")

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
		cleanupConfigEnv(t)
		t.Cleanup(func() { cleanupConfigEnv(t) })

		os.Setenv("MACROLENS_SERVER_PORT", "9090")
		os.Setenv("MACROLENS_SERVER_ENVIRONMENT", "production")
		os.Setenv("MACROLENS_SERVER_ALLOWED_ORIGINS", "http://localhost:3000,https://example.com")
		os.Setenv("MACROLENS_USDA_API_KEY", "custom-api-key")
		os.Setenv("MACROLENS_USDA_BASE_URL", "https://custom.api.com")
		os.Setenv("MACROLENS_CACHE_TYPE", "redis")
		os.Setenv("MACROLENS_CACHE_REDIS_URL", "redis://localhost:6379")
		os.Setenv("MACROLENS_CACHE_TTL", "24h")
		os.Setenv("MACROLENS_RATELIMIT_PER_IP", "200")
		os.Setenv("MACROLENS_RATELIMIT_USDA", "2000")

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
		if len(cfg.Server.AllowedOrigins) != 2 {
			t.Errorf("Server.AllowedOrigins length = %d, want 2", len(cfg.Server.AllowedOrigins))
		}
		if len(cfg.Server.AllowedOrigins) >= 2 {
			if cfg.Server.AllowedOrigins[0] != "http://localhost:3000" {
				t.Errorf("Server.AllowedOrigins[0] = %s, want http://localhost:3000", cfg.Server.AllowedOrigins[0])
			}
			if cfg.Server.AllowedOrigins[1] != "https://example.com" {
				t.Errorf("Server.AllowedOrigins[1] = %s, want https://example.com", cfg.Server.AllowedOrigins[1])
			}
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
		cleanupConfigEnv(t)
		t.Cleanup(func() { cleanupConfigEnv(t) })

		_, err := Load()
		if err == nil {
			t.Error("Load() error = nil, want error for missing API key")
		}
		if err != nil && err.Error() != "invalid configuration: USDA API key is required (set MACROLENS_USDA_API_KEY)" {
			t.Errorf("Load() error = %v, want 'USDA API key is required'", err)
		}
	})

	t.Run("fails validation for invalid cache type", func(t *testing.T) {
		cleanupConfigEnv(t)
		t.Cleanup(func() { cleanupConfigEnv(t) })

		os.Setenv("MACROLENS_USDA_API_KEY", "test-key")
		os.Setenv("MACROLENS_CACHE_TYPE", "invalid")

		_, err := Load()
		if err == nil {
			t.Error("Load() error = nil, want error for invalid cache type")
		}
	})

	t.Run("fails validation when redis URL missing for redis cache", func(t *testing.T) {
		cleanupConfigEnv(t)
		t.Cleanup(func() { cleanupConfigEnv(t) })

		os.Setenv("MACROLENS_USDA_API_KEY", "test-key")
		os.Setenv("MACROLENS_CACHE_TYPE", "redis")

		_, err := Load()
		if err == nil {
			t.Error("Load() error = nil, want error for missing Redis URL")
		}
	})

	t.Run("loads single allowed origin from environment variable", func(t *testing.T) {
		cleanupConfigEnv(t)
		t.Cleanup(func() { cleanupConfigEnv(t) })

		os.Setenv("MACROLENS_USDA_API_KEY", "test-key")
		os.Setenv("MACROLENS_SERVER_ALLOWED_ORIGINS", "chrome-extension://*")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}

		if len(cfg.Server.AllowedOrigins) != 1 {
			t.Errorf("Server.AllowedOrigins length = %d, want 1", len(cfg.Server.AllowedOrigins))
		}
		if len(cfg.Server.AllowedOrigins) >= 1 && cfg.Server.AllowedOrigins[0] != "chrome-extension://*" {
			t.Errorf("Server.AllowedOrigins[0] = %s, want chrome-extension://*", cfg.Server.AllowedOrigins[0])
		}
	})

	t.Run("loads multiple allowed origins with spaces from environment variable", func(t *testing.T) {
		cleanupConfigEnv(t)
		t.Cleanup(func() { cleanupConfigEnv(t) })

		os.Setenv("MACROLENS_USDA_API_KEY", "test-key")
		os.Setenv("MACROLENS_SERVER_ALLOWED_ORIGINS", "http://localhost:3000, https://example.com, chrome-extension://*")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}

		if len(cfg.Server.AllowedOrigins) != 3 {
			t.Errorf("Server.AllowedOrigins length = %d, want 3", len(cfg.Server.AllowedOrigins))
		}
		// Viper may trim spaces, so we check for both possibilities
		expectedOrigins := []string{"http://localhost:3000", "https://example.com", "chrome-extension://*"}
		for i, expected := range expectedOrigins {
			if len(cfg.Server.AllowedOrigins) > i {
				actual := cfg.Server.AllowedOrigins[i]
				// Remove any whitespace that might be present
				if actual != expected && strings.TrimSpace(actual) != expected {
					t.Errorf("Server.AllowedOrigins[%d] = %q, want %q", i, actual, expected)
				}
			}
		}
	})

	t.Run("uses default allowed origins when not set", func(t *testing.T) {
		cleanupConfigEnv(t)
		t.Cleanup(func() { cleanupConfigEnv(t) })

		os.Setenv("MACROLENS_USDA_API_KEY", "test-key")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}

		if len(cfg.Server.AllowedOrigins) != 1 {
			t.Errorf("Server.AllowedOrigins length = %d, want 1 (default)", len(cfg.Server.AllowedOrigins))
		}
		if len(cfg.Server.AllowedOrigins) >= 1 && cfg.Server.AllowedOrigins[0] != "chrome-extension://*" {
			t.Errorf("Server.AllowedOrigins[0] = %s, want chrome-extension://* (default)", cfg.Server.AllowedOrigins[0])
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

	t.Run("handles Windows-style line endings (CRLF)", func(t *testing.T) {
		// Save current directory
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create temp directory
		tempDir := t.TempDir()
		os.Chdir(tempDir)

		// Create .env file with Windows line endings (\r\n)
		envContent := "TEST_WINDOWS_1=value1\r\nTEST_WINDOWS_2=value2\r\n# Comment\r\nTEST_WINDOWS_3=value3\r\n"
		err := os.WriteFile(".env", []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}

		os.Unsetenv("TEST_WINDOWS_1")
		os.Unsetenv("TEST_WINDOWS_2")
		os.Unsetenv("TEST_WINDOWS_3")

		err = loadEnvFile()
		if err != nil {
			t.Fatalf("loadEnvFile() error = %v, want nil", err)
		}

		if os.Getenv("TEST_WINDOWS_1") != "value1" {
			t.Errorf("TEST_WINDOWS_1 = %q, want 'value1'", os.Getenv("TEST_WINDOWS_1"))
		}
		if os.Getenv("TEST_WINDOWS_2") != "value2" {
			t.Errorf("TEST_WINDOWS_2 = %q, want 'value2'", os.Getenv("TEST_WINDOWS_2"))
		}
		if os.Getenv("TEST_WINDOWS_3") != "value3" {
			t.Errorf("TEST_WINDOWS_3 = %q, want 'value3'", os.Getenv("TEST_WINDOWS_3"))
		}

		os.Unsetenv("TEST_WINDOWS_1")
		os.Unsetenv("TEST_WINDOWS_2")
		os.Unsetenv("TEST_WINDOWS_3")
	})

	t.Run("handles mixed Unix and Windows line endings", func(t *testing.T) {
		// Save current directory
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		// Create temp directory
		tempDir := t.TempDir()
		os.Chdir(tempDir)

		// Create .env file with mixed line endings
		envContent := "TEST_MIXED_1=value1\r\nTEST_MIXED_2=value2\nTEST_MIXED_3=value3\r\n"
		err := os.WriteFile(".env", []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}

		os.Unsetenv("TEST_MIXED_1")
		os.Unsetenv("TEST_MIXED_2")
		os.Unsetenv("TEST_MIXED_3")

		err = loadEnvFile()
		if err != nil {
			t.Fatalf("loadEnvFile() error = %v, want nil", err)
		}

		if os.Getenv("TEST_MIXED_1") != "value1" {
			t.Errorf("TEST_MIXED_1 = %q, want 'value1'", os.Getenv("TEST_MIXED_1"))
		}
		if os.Getenv("TEST_MIXED_2") != "value2" {
			t.Errorf("TEST_MIXED_2 = %q, want 'value2'", os.Getenv("TEST_MIXED_2"))
		}
		if os.Getenv("TEST_MIXED_3") != "value3" {
			t.Errorf("TEST_MIXED_3 = %q, want 'value3'", os.Getenv("TEST_MIXED_3"))
		}

		os.Unsetenv("TEST_MIXED_1")
		os.Unsetenv("TEST_MIXED_2")
		os.Unsetenv("TEST_MIXED_3")
	})

	t.Run("strips double quotes from values", func(t *testing.T) {
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		tempDir := t.TempDir()
		os.Chdir(tempDir)

		envContent := `TEST_QUOTED_DOUBLE="value with spaces"
TEST_QUOTED_SPECIAL="value=with=equals"
TEST_NOT_QUOTED=plain_value`
		err := os.WriteFile(".env", []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}

		os.Unsetenv("TEST_QUOTED_DOUBLE")
		os.Unsetenv("TEST_QUOTED_SPECIAL")
		os.Unsetenv("TEST_NOT_QUOTED")

		err = loadEnvFile()
		if err != nil {
			t.Fatalf("loadEnvFile() error = %v, want nil", err)
		}

		if os.Getenv("TEST_QUOTED_DOUBLE") != "value with spaces" {
			t.Errorf("TEST_QUOTED_DOUBLE = %q, want 'value with spaces'", os.Getenv("TEST_QUOTED_DOUBLE"))
		}
		if os.Getenv("TEST_QUOTED_SPECIAL") != "value=with=equals" {
			t.Errorf("TEST_QUOTED_SPECIAL = %q, want 'value=with=equals'", os.Getenv("TEST_QUOTED_SPECIAL"))
		}
		if os.Getenv("TEST_NOT_QUOTED") != "plain_value" {
			t.Errorf("TEST_NOT_QUOTED = %q, want 'plain_value'", os.Getenv("TEST_NOT_QUOTED"))
		}

		os.Unsetenv("TEST_QUOTED_DOUBLE")
		os.Unsetenv("TEST_QUOTED_SPECIAL")
		os.Unsetenv("TEST_NOT_QUOTED")
	})

	t.Run("strips single quotes from values", func(t *testing.T) {
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		tempDir := t.TempDir()
		os.Chdir(tempDir)

		envContent := "TEST_SINGLE_QUOTED='value with spaces'\nTEST_SINGLE_SPECIAL='value#with#hash'"
		err := os.WriteFile(".env", []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}

		os.Unsetenv("TEST_SINGLE_QUOTED")
		os.Unsetenv("TEST_SINGLE_SPECIAL")

		err = loadEnvFile()
		if err != nil {
			t.Fatalf("loadEnvFile() error = %v, want nil", err)
		}

		if os.Getenv("TEST_SINGLE_QUOTED") != "value with spaces" {
			t.Errorf("TEST_SINGLE_QUOTED = %q, want 'value with spaces'", os.Getenv("TEST_SINGLE_QUOTED"))
		}
		if os.Getenv("TEST_SINGLE_SPECIAL") != "value#with#hash" {
			t.Errorf("TEST_SINGLE_SPECIAL = %q, want 'value#with#hash'", os.Getenv("TEST_SINGLE_SPECIAL"))
		}

		os.Unsetenv("TEST_SINGLE_QUOTED")
		os.Unsetenv("TEST_SINGLE_SPECIAL")
	})

	t.Run("warns about malformed lines", func(t *testing.T) {
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		tempDir := t.TempDir()
		os.Chdir(tempDir)

		// Create .env with malformed line (no equals sign)
		envContent := `TEST_VALID=value1
MALFORMED_LINE_NO_EQUALS
TEST_VALID_2=value2`
		err := os.WriteFile(".env", []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}

		os.Unsetenv("TEST_VALID")
		os.Unsetenv("TEST_VALID_2")
		os.Unsetenv("MALFORMED_LINE_NO_EQUALS")

		// Capture stderr to check for warning
		err = loadEnvFile()
		if err != nil {
			t.Fatalf("loadEnvFile() error = %v, want nil (should not fail on malformed lines)", err)
		}

		// Valid lines should be loaded
		if os.Getenv("TEST_VALID") != "value1" {
			t.Errorf("TEST_VALID = %q, want 'value1'", os.Getenv("TEST_VALID"))
		}
		if os.Getenv("TEST_VALID_2") != "value2" {
			t.Errorf("TEST_VALID_2 = %q, want 'value2'", os.Getenv("TEST_VALID_2"))
		}

		// Malformed line should not create an env var
		if os.Getenv("MALFORMED_LINE_NO_EQUALS") != "" {
			t.Errorf("MALFORMED_LINE_NO_EQUALS should not be set, got %q", os.Getenv("MALFORMED_LINE_NO_EQUALS"))
		}

		os.Unsetenv("TEST_VALID")
		os.Unsetenv("TEST_VALID_2")
	})
}

func TestUnquoteValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "double quoted value",
			input: `"hello world"`,
			want:  "hello world",
		},
		{
			name:  "single quoted value",
			input: "'hello world'",
			want:  "hello world",
		},
		{
			name:  "no quotes",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only opening quote",
			input: `"hello`,
			want:  `"hello`,
		},
		{
			name:  "only closing quote",
			input: `hello"`,
			want:  `hello"`,
		},
		{
			name:  "mismatched quotes",
			input: `"hello'`,
			want:  `"hello'`,
		},
		{
			name:  "empty quoted string",
			input: `""`,
			want:  "",
		},
		{
			name:  "single character",
			input: "a",
			want:  "a",
		},
		{
			name:  "value with equals sign",
			input: `"key=value"`,
			want:  "key=value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unquoteValue(tt.input)
			if got != tt.want {
				t.Errorf("unquoteValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
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
