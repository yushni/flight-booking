package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Providers ProvidersConfig `mapstructure:"providers"`
	Cache     CacheConfig     `mapstructure:"cache"`
	Log       LogConfig       `mapstructure:"log"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port                    string        `mapstructure:"port"`
	Host                    string        `mapstructure:"host"`
	ReadTimeout             time.Duration `mapstructure:"read_timeout"`
	WriteTimeout            time.Duration `mapstructure:"write_timeout"`
	IdleTimeout             time.Duration `mapstructure:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
}

// ProvidersConfig represents providers configuration
type ProvidersConfig struct {
	Provider1 ProviderConfig `mapstructure:"provider1"`
	Provider2 ProviderConfig `mapstructure:"provider2"`
}

// ProviderConfig represents individual provider configuration
type ProviderConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	BaseURL string        `mapstructure:"base_url"`
	Timeout time.Duration `mapstructure:"timeout"`
	Retries int           `mapstructure:"retries"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	DefaultTTL      time.Duration `mapstructure:"default_ttl"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
	MaxSize         int           `mapstructure:"max_size"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Configure viper
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/flight-booking")

	// Enable reading from environment variables
	v.AutomaticEnv()

	// Set environment variable prefix
	v.SetEnvPrefix("FLIGHT_BOOKING")

	// Read configuration file (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal configuration
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.host", "")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "60s")
	v.SetDefault("server.graceful_shutdown_timeout", "30s")

	// Provider 1 defaults
	v.SetDefault("providers.provider1.enabled", true)
	v.SetDefault("providers.provider1.base_url", "https://api.provider1.com")
	v.SetDefault("providers.provider1.timeout", "10s")
	v.SetDefault("providers.provider1.retries", 3)

	// Provider 2 defaults
	v.SetDefault("providers.provider2.enabled", true)
	v.SetDefault("providers.provider2.base_url", "https://api.provider2.com")
	v.SetDefault("providers.provider2.timeout", "10s")
	v.SetDefault("providers.provider2.retries", 3)

	// Cache defaults
	v.SetDefault("cache.enabled", true)
	v.SetDefault("cache.default_ttl", "5m")
	v.SetDefault("cache.cleanup_interval", "10m")
	v.SetDefault("cache.max_size", 10000)

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate server configuration
	if config.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if config.Server.ReadTimeout <= 0 {
		return fmt.Errorf("server read timeout must be positive")
	}

	if config.Server.WriteTimeout <= 0 {
		return fmt.Errorf("server write timeout must be positive")
	}

	// Validate providers configuration
	if !config.Providers.Provider1.Enabled && !config.Providers.Provider2.Enabled {
		return fmt.Errorf("at least one provider must be enabled")
	}

	if config.Providers.Provider1.Enabled && config.Providers.Provider1.BaseURL == "" {
		return fmt.Errorf("provider1 base URL is required when enabled")
	}

	if config.Providers.Provider2.Enabled && config.Providers.Provider2.BaseURL == "" {
		return fmt.Errorf("provider2 base URL is required when enabled")
	}

	// Validate cache configuration
	if config.Cache.Enabled && config.Cache.DefaultTTL <= 0 {
		return fmt.Errorf("cache default TTL must be positive when cache is enabled")
	}

	// Validate log configuration
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[config.Log.Level] {
		return fmt.Errorf("invalid log level: %s", config.Log.Level)
	}

	validLogFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if !validLogFormats[config.Log.Format] {
		return fmt.Errorf("invalid log format: %s", config.Log.Format)
	}

	return nil
}
