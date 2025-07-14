package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"flight-booking/internal/config"
	"flight-booking/internal/interfaces"
	"flight-booking/internal/models"
)

// BaseProvider provides common functionality for all providers
type BaseProvider struct {
	name       string
	config     config.ProviderConfig
	httpClient interfaces.HTTPClient
	cache      interfaces.CacheService
	logger     interfaces.Logger
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(
	name string,
	config config.ProviderConfig,
	httpClient interfaces.HTTPClient,
	cache interfaces.CacheService,
	logger interfaces.Logger,
) *BaseProvider {
	return &BaseProvider{
		name:       name,
		config:     config,
		httpClient: httpClient,
		cache:      cache,
		logger:     logger.With("provider", name),
	}
}

// GetName returns the provider name
func (p *BaseProvider) GetName() string {
	return p.name
}

// IsHealthy checks if the provider is healthy
func (p *BaseProvider) IsHealthy(ctx context.Context) error {
	if !p.config.Enabled {
		return fmt.Errorf("provider %s is disabled", p.name)
	}

	// Create a simple health check URL
	healthURL := fmt.Sprintf("%s/health", p.config.BaseURL)

	// Try to make a request with a shorter timeout
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := p.httpClient.Get(healthCtx, healthURL, nil)
	if err != nil {
		return fmt.Errorf("provider %s health check failed: %w", p.name, err)
	}

	return nil
}

// buildURL builds a URL with query parameters
func (p *BaseProvider) buildURL(endpoint string, filters models.RouteFilters) string {
	u, _ := url.Parse(p.config.BaseURL + endpoint)

	query := u.Query()
	if filters.Airline != "" {
		query.Set("airline", filters.Airline)
	}
	if filters.SourceAirport != "" {
		query.Set("source", filters.SourceAirport)
	}
	if filters.DestinationAirport != "" {
		query.Set("destination", filters.DestinationAirport)
	}
	if filters.MaxStops != nil {
		query.Set("maxStops", fmt.Sprintf("%d", *filters.MaxStops))
	}

	u.RawQuery = query.Encode()
	return u.String()
}

// makeRequest makes an HTTP request with retries
func (p *BaseProvider) makeRequest(ctx context.Context, url string) ([]byte, error) {
	var lastErr error

	for i := 0; i < p.config.Retries; i++ {
		if i > 0 {
			// Add exponential backoff
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i) * time.Second):
			}
		}

		data, err := p.httpClient.Get(ctx, url, nil)
		if err == nil {
			return data, nil
		}

		lastErr = err
		p.logger.Warn(fmt.Sprintf("Request failed (attempt %d/%d)", i+1, p.config.Retries), "error", err)
	}

	return nil, fmt.Errorf("all retry attempts failed: %w", lastErr)
}

// getCachedRoutes retrieves routes from cache
func (p *BaseProvider) getCachedRoutes(cacheKey string) ([]models.FlightRoute, bool) {
	if data, found := p.cache.Get(cacheKey); found {
		if routes, ok := data.([]models.FlightRoute); ok {
			p.logger.Debug("Cache hit", "key", cacheKey, "routes_count", len(routes))
			return routes, true
		}
	}
	return nil, false
}

// setCachedRoutes stores routes in cache
func (p *BaseProvider) setCachedRoutes(cacheKey string, routes []models.FlightRoute, ttl time.Duration) {
	p.cache.Set(cacheKey, routes, ttl)
	p.logger.Debug("Cache set", "key", cacheKey, "routes_count", len(routes), "ttl", ttl)
}

// parseRoutes parses the response and converts it to FlightRoute objects
func (p *BaseProvider) parseRoutes(data []byte) ([]models.FlightRoute, error) {
	var response struct {
		Routes []models.FlightRoute `json:"routes"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		// Try to parse as direct array
		var routes []models.FlightRoute
		if err := json.Unmarshal(data, &routes); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		return routes, nil
	}

	return response.Routes, nil
}

// validateAndProcessRoutes validates and processes routes
func (p *BaseProvider) validateAndProcessRoutes(routes []models.FlightRoute) []models.FlightRoute {
	validRoutes := make([]models.FlightRoute, 0, len(routes))

	for _, route := range routes {
		if err := route.Validate(); err != nil {
			p.logger.Warn("Invalid route skipped", "error", err, "route", route)
			continue
		}

		// Set provider name
		route.Provider = p.name
		validRoutes = append(validRoutes, route)
	}

	return validRoutes
}
