package providers

import (
	"context"
	"fmt"
	"time"

	"flight-booking/internal/config"
	"flight-booking/internal/interfaces"
	"flight-booking/internal/models"
)

// Provider1 implements the RouteProvider interface for Provider1
type Provider1 struct {
	*BaseProvider
}

// NewProvider1 creates a new Provider1 instance
func NewProvider1(
	config config.ProviderConfig,
	httpClient interfaces.HTTPClient,
	cache interfaces.CacheService,
	logger interfaces.Logger,
) interfaces.RouteProvider {
	baseProvider := NewBaseProvider("provider1", config, httpClient, cache, logger)
	return &Provider1{
		BaseProvider: baseProvider,
	}
}

// GetRoutes fetches flight routes from Provider1
func (p *Provider1) GetRoutes(ctx context.Context, filters models.RouteFilters) ([]models.FlightRoute, error) {
	if !p.config.Enabled {
		p.logger.Info("Provider1 is disabled")
		return []models.FlightRoute{}, nil
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("provider1:%s", models.CacheKey{Filters: filters}.String())

	// Check cache first
	if cachedRoutes, found := p.getCachedRoutes(cacheKey); found {
		return cachedRoutes, nil
	}

	// Build request URL
	url := p.buildURL("/routes", filters)
	p.logger.Debug("Fetching routes from Provider1", "url", url)

	// Make request with timeout
	requestCtx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	defer cancel()

	data, err := p.makeRequest(requestCtx, url)
	if err != nil {
		p.logger.Error("Failed to fetch routes from Provider1", "error", err)
		return nil, fmt.Errorf("provider1 request failed: %w", err)
	}

	// Parse response
	routes, err := p.parseRoutes(data)
	if err != nil {
		p.logger.Error("Failed to parse Provider1 response", "error", err)
		return nil, fmt.Errorf("provider1 parse failed: %w", err)
	}

	// Validate and process routes
	validRoutes := p.validateAndProcessRoutes(routes)

	// Apply additional Provider1-specific processing
	processedRoutes := p.processProvider1Routes(validRoutes, filters)

	// Cache the results
	p.setCachedRoutes(cacheKey, processedRoutes, 5*time.Minute)

	p.logger.Info("Successfully fetched routes from Provider1",
		"total_routes", len(processedRoutes),
		"cached_key", cacheKey)

	return processedRoutes, nil
}

// processProvider1Routes applies Provider1-specific processing
func (p *Provider1) processProvider1Routes(routes []models.FlightRoute, filters models.RouteFilters) []models.FlightRoute {
	// Apply any Provider1-specific transformations
	for i := range routes {
		// Provider1 might have specific data formatting
		routes[i].Provider = "provider1"

		// Provider1 might need specific equipment mapping
		if routes[i].Equipment != nil && *routes[i].Equipment == "UNKNOWN" {
			routes[i].Equipment = nil
		}
	}

	// Apply filters
	return filters.ApplyFilters(routes)
}

// IsHealthy checks if Provider1 is healthy
func (p *Provider1) IsHealthy(ctx context.Context) error {
	if !p.config.Enabled {
		return fmt.Errorf("provider1 is disabled")
	}

	// Use base provider health check
	return p.BaseProvider.IsHealthy(ctx)
}
