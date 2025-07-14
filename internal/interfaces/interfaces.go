package interfaces

import (
	"context"
	"time"

	"flight-booking/internal/models"
)

// RouteProvider defines the interface for route providers
type RouteProvider interface {
	GetRoutes(ctx context.Context, filters models.RouteFilters) ([]models.FlightRoute, error)
	GetName() string
	IsHealthy(ctx context.Context) error
}

// CacheService defines the interface for cache operations
type CacheService interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
}

// RouteService defines the interface for route business logic
type RouteService interface {
	GetRoutes(ctx context.Context, filters models.RouteFilters) (*models.RoutesResponse, error)
	GetHealth(ctx context.Context) (*models.HealthResponse, error)
}

// Logger defines the interface for logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	With(fields ...interface{}) Logger
}

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	Get(ctx context.Context, url string, headers map[string]string) ([]byte, error)
	Post(ctx context.Context, url string, body []byte, headers map[string]string) ([]byte, error)
	Put(ctx context.Context, url string, body []byte, headers map[string]string) ([]byte, error)
	Delete(ctx context.Context, url string, headers map[string]string) ([]byte, error)
}

// MetricsCollector defines the interface for metrics collection
type MetricsCollector interface {
	IncrementCounter(name string, tags map[string]string)
	RecordGauge(name string, value float64, tags map[string]string)
	RecordHistogram(name string, value float64, tags map[string]string)
	RecordTiming(name string, duration time.Duration, tags map[string]string)
}

// ProviderManager defines the interface for managing multiple providers
type ProviderManager interface {
	GetAllProviders() []RouteProvider
	GetEnabledProviders() []RouteProvider
	GetProvider(name string) (RouteProvider, bool)
}

// ConfigProvider defines the interface for configuration access
type ConfigProvider interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetDuration(key string) time.Duration
}
