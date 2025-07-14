package providers

import (
	"context"
	"fmt"
	"time"

	"flight-booking/internal/config"
	"flight-booking/internal/interfaces"
	"flight-booking/internal/models"
)

// Provider2 implements the RouteProvider interface for Provider2
type Provider2 struct {
	*BaseProvider
}

// NewProvider2 creates a new Provider2 instance
func NewProvider2(
	config config.ProviderConfig,
	httpClient interfaces.HTTPClient,
	cache interfaces.CacheService,
	logger interfaces.Logger,
) interfaces.RouteProvider {
	baseProvider := NewBaseProvider("provider2", config, httpClient, cache, logger)
	return &Provider2{
		BaseProvider: baseProvider,
	}
}

// GetRoutes fetches flight routes from Provider2
func (p *Provider2) GetRoutes(ctx context.Context, filters models.RouteFilters) ([]models.FlightRoute, error) {
	if !p.config.Enabled {
		p.logger.Info("Provider2 is disabled")
		return []models.FlightRoute{}, nil
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("provider2:%s", models.CacheKey{Filters: filters}.String())

	// Check cache first
	if cachedRoutes, found := p.getCachedRoutes(cacheKey); found {
		return cachedRoutes, nil
	}

	// Build request URL
	url := p.buildURL("/routes", filters)
	p.logger.Debug("Fetching routes from Provider2", "url", url)

	// Make request with timeout
	requestCtx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	defer cancel()

	data, err := p.makeRequest(requestCtx, url)
	if err != nil {
		p.logger.Error("Failed to fetch routes from Provider2", "error", err)
		return nil, fmt.Errorf("provider2 request failed: %w", err)
	}

	// Parse response
	routes, err := p.parseRoutes(data)
	if err != nil {
		p.logger.Error("Failed to parse Provider2 response", "error", err)
		return nil, fmt.Errorf("provider2 parse failed: %w", err)
	}

	// Validate and process routes
	validRoutes := p.validateAndProcessRoutes(routes)

	// Apply additional Provider2-specific processing
	processedRoutes := p.processProvider2Routes(validRoutes, filters)

	// Cache the results
	p.setCachedRoutes(cacheKey, processedRoutes, 5*time.Minute)

	p.logger.Info("Successfully fetched routes from Provider2",
		"total_routes", len(processedRoutes),
		"cached_key", cacheKey)

	return processedRoutes, nil
}

// processProvider2Routes applies Provider2-specific processing
func (p *Provider2) processProvider2Routes(routes []models.FlightRoute, filters models.RouteFilters) []models.FlightRoute {
	// Apply any Provider2-specific transformations
	for i := range routes {
		// Provider2 might have specific data formatting
		routes[i].Provider = "provider2"

		// Provider2 might have different equipment format
		if routes[i].Equipment != nil && *routes[i].Equipment == "" {
			routes[i].Equipment = nil
		}
	}

	// Apply filters
	return filters.ApplyFilters(routes)
}

// IsHealthy checks if Provider2 is healthy
func (p *Provider2) IsHealthy(ctx context.Context) error {
	if !p.config.Enabled {
		return fmt.Errorf("provider2 is disabled")
	}

	// Use base provider health check
	return p.BaseProvider.IsHealthy(ctx)
}
