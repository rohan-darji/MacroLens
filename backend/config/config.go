package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	USDA      USDAConfig
	Cache     CacheConfig
	RateLimit RateLimitConfig
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

	// Set config name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/macrolens/")

	// Environment variable settings
	v.SetEnvPrefix("MACROLENS")
	v.AutomaticEnv()

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
