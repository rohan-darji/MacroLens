package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	USDA      USDAConfig
	Cache     CacheConfig
	RateLimit RateLimitConfig
	Matching  MatchingConfig
}

// MatchingConfig holds product matching algorithm configuration
type MatchingConfig struct {
	MinConfidenceThreshold float64 `mapstructure:"min_confidence_threshold"`
	EnableFuzzyMatching    bool    `mapstructure:"enable_fuzzy_matching"`
	EnableDebugLogging     bool    `mapstructure:"enable_debug_logging"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port            string   `mapstructure:"port"`
	Environment     string   `mapstructure:"environment"`
	AllowedOrigins  []string `mapstructure:"allowed_origins"`
}

// USDAConfig holds USDA API configuration
type USDAConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
}

// CacheConfig holds cache-related configuration
type CacheConfig struct {
	Type      string        `mapstructure:"type"` // "memory" or "redis"
	RedisURL  string        `mapstructure:"redis_url"`
	TTL       time.Duration `mapstructure:"ttl"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	PerIP int `mapstructure:"per_ip"`
	USDA  int `mapstructure:"usda"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	v := viper.New()

	// Load .env file if it exists
	if err := loadEnvFile(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	// Set config name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/macrolens/")

	// Environment variable settings
	v.SetEnvPrefix("MACROLENS")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific environment variables to config keys
	bindEnvVars(v)

	// Set default values
	setDefaults(v)

	// Read config file (optional - will use env vars if file doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; using environment variables and defaults
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// loadEnvFile loads .env file into environment variables
func loadEnvFile() error {
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		// .env file doesn't exist, that's okay
		return nil
	}

	data, err := os.ReadFile(envFile)
	if err != nil {
		return err
	}

	// Handle both Unix (\n) and Windows (\r\n) line endings
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Warn about malformed lines (not empty, not comment, but no '=')
			fmt.Fprintf(os.Stderr, "warning: ignoring malformed line %d in .env: %q\n", lineNum+1, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Strip surrounding quotes from value (supports both single and double quotes)
		value = unquoteValue(value)

		// Only set if not already in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return nil
}

// unquoteValue removes surrounding quotes from a value
// Supports both double quotes (") and single quotes (')
func unquoteValue(value string) string {
	if len(value) >= 2 {
		// Check for double quotes
		if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
			return value[1 : len(value)-1]
		}
		// Check for single quotes
		if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
			return value[1 : len(value)-1]
		}
	}
	return value
}

// bindEnvVars binds environment variables to config keys
func bindEnvVars(v *viper.Viper) {
	// Server
	v.BindEnv("server.port", "MACROLENS_SERVER_PORT")
	v.BindEnv("server.environment", "MACROLENS_SERVER_ENVIRONMENT")
	v.BindEnv("server.allowed_origins", "MACROLENS_SERVER_ALLOWED_ORIGINS")

	// USDA
	v.BindEnv("usda.api_key", "MACROLENS_USDA_API_KEY")
	v.BindEnv("usda.base_url", "MACROLENS_USDA_BASE_URL")

	// Cache
	v.BindEnv("cache.type", "MACROLENS_CACHE_TYPE")
	v.BindEnv("cache.redis_url", "MACROLENS_CACHE_REDIS_URL")
	v.BindEnv("cache.ttl", "MACROLENS_CACHE_TTL")

	// Rate Limit
	v.BindEnv("ratelimit.per_ip", "MACROLENS_RATELIMIT_PER_IP")
	v.BindEnv("ratelimit.usda", "MACROLENS_RATELIMIT_USDA")

	// Matching
	v.BindEnv("matching.min_confidence_threshold", "MACROLENS_MATCHING_MIN_CONFIDENCE")
	v.BindEnv("matching.enable_fuzzy_matching", "MACROLENS_MATCHING_ENABLE_FUZZY")
	v.BindEnv("matching.enable_debug_logging", "MACROLENS_MATCHING_DEBUG")
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.environment", "development")
	v.SetDefault("server.allowed_origins", []string{"chrome-extension://*"})

	// USDA defaults
	v.SetDefault("usda.base_url", "https://api.nal.usda.gov/fdc")

	// Cache defaults
	v.SetDefault("cache.type", "memory")
	v.SetDefault("cache.ttl", "720h") // 30 days

	// Rate limit defaults
	v.SetDefault("ratelimit.per_ip", 100)
	v.SetDefault("ratelimit.usda", 1000)

	// Matching defaults
	v.SetDefault("matching.min_confidence_threshold", 40.0)
	v.SetDefault("matching.enable_fuzzy_matching", true)
	v.SetDefault("matching.enable_debug_logging", false)
}

// validate validates the configuration
func validate(config *Config) error {
	if config.USDA.APIKey == "" {
		return fmt.Errorf("USDA API key is required (set MACROLENS_USDA_API_KEY)")
	}

	if config.Cache.Type != "memory" && config.Cache.Type != "redis" {
		return fmt.Errorf("cache type must be 'memory' or 'redis', got: %s", config.Cache.Type)
	}

	if config.Cache.Type == "redis" && config.Cache.RedisURL == "" {
		return fmt.Errorf("Redis URL is required when cache type is 'redis'")
	}

	return nil
}
