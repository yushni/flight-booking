package config

import (
	"time"
)

// Config represents the application configuration
type Config struct {
	AppShutdownTimeout time.Duration `mapstructure:"app_shutdown_timeout"`

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
	Providers []ProviderConfig `json:"providers" yaml:"providers"`
	CacheTTL  time.Duration    `json:"cacheTtl" yaml:"cacheTtl"`
}

// ProviderConfig represents individual provider configuration
type ProviderConfig struct {
	Name    string        `mapstructure:"name"`
	Enabled bool          `mapstructure:"enabled"`
	URL     string        `mapstructure:"url"`
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
