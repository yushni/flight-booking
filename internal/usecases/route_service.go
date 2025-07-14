package usecases

import (
	"context"
	"fmt"
	"sort"
	"time"

	"flight-booking/internal/interfaces"
	"flight-booking/internal/models"
	"flight-booking/internal/providers"
)

// RouteServiceImpl implements the RouteService interface
type RouteServiceImpl struct {
	providerManager interfaces.ProviderManager
	cache           interfaces.CacheService
	logger          interfaces.Logger
}

// NewRouteService creates a new route service
func NewRouteService(
	providerManager interfaces.ProviderManager,
	cache interfaces.CacheService,
	logger interfaces.Logger,
) interfaces.RouteService {
	return &RouteServiceImpl{
		providerManager: providerManager,
		cache:           cache,
		logger:          logger.With("component", "route_service"),
	}
}

// GetRoutes aggregates routes from all providers
func (rs *RouteServiceImpl) GetRoutes(ctx context.Context, filters models.RouteFilters) (*models.RoutesResponse, error) {
	rs.logger.Info("Fetching routes", "filters", filters)

	// Generate cache key for aggregated results
	cacheKey := fmt.Sprintf("aggregated:%s", models.CacheKey{Filters: filters}.String())

	// Check cache first
	if cachedData, found := rs.cache.Get(cacheKey); found {
		if response, ok := cachedData.(*models.RoutesResponse); ok {
			response.Metadata.CacheHit = true
			response.Metadata.Timestamp = time.Now()
			rs.logger.Debug("Cache hit for aggregated routes", "key", cacheKey)
			return response, nil
		}
	}

	// Get provider manager interface
	var pm *providers.ProviderManager
	if pmTyped, ok := rs.providerManager.(*providers.ProviderManager); ok {
		pm = pmTyped
	} else {
		// If we can't cast, we'll use the interface methods
		return rs.getRoutesFromInterface(ctx, filters, cacheKey)
	}

	// Fetch routes from all providers concurrently
	providerResponses := pm.GetRoutesFromAllProviders(ctx, filters)

	// Aggregate results
	aggregatedRoutes := make([]models.FlightRoute, 0)
	providersUsed := make([]string, 0)

	for _, response := range providerResponses {
		if response.Error != nil {
			rs.logger.Error("Provider error",
				"provider", response.Provider,
				"error", response.Error)
			continue
		}

		aggregatedRoutes = append(aggregatedRoutes, response.Routes...)
		providersUsed = append(providersUsed, response.Provider)
	}

	// Remove duplicates and sort
	uniqueRoutes := rs.removeDuplicates(aggregatedRoutes)
	sortedRoutes := rs.sortRoutes(uniqueRoutes)

	// Create response
	response := &models.RoutesResponse{
		Data: sortedRoutes,
		Metadata: models.ResponseMetadata{
			TotalCount:    len(sortedRoutes),
			ProvidersUsed: providersUsed,
			CacheHit:      false,
			Timestamp:     time.Now(),
		},
	}

	// Cache the response
	rs.cache.Set(cacheKey, response, 5*time.Minute)

	rs.logger.Info("Successfully aggregated routes",
		"total_routes", len(sortedRoutes),
		"providers_used", len(providersUsed),
		"cache_key", cacheKey)

	return response, nil
}

// getRoutesFromInterface is a fallback method using interface methods
func (rs *RouteServiceImpl) getRoutesFromInterface(ctx context.Context, filters models.RouteFilters, cacheKey string) (*models.RoutesResponse, error) {
	enabledProviders := rs.providerManager.GetEnabledProviders()

	if len(enabledProviders) == 0 {
		rs.logger.Warn("No enabled providers found")
		return &models.RoutesResponse{
			Data: []models.FlightRoute{},
			Metadata: models.ResponseMetadata{
				TotalCount:    0,
				ProvidersUsed: []string{},
				CacheHit:      false,
				Timestamp:     time.Now(),
			},
		}, nil
	}

	// Create channels for concurrent execution
	type providerResult struct {
		routes   []models.FlightRoute
		provider string
		err      error
	}

	resultChannel := make(chan providerResult, len(enabledProviders))

	// Launch goroutines for each provider
	for _, provider := range enabledProviders {
		go func(p interfaces.RouteProvider) {
			routes, err := p.GetRoutes(ctx, filters)
			resultChannel <- providerResult{
				routes:   routes,
				provider: p.GetName(),
				err:      err,
			}
		}(provider)
	}

	// Collect results
	aggregatedRoutes := make([]models.FlightRoute, 0)
	providersUsed := make([]string, 0)

	for i := 0; i < len(enabledProviders); i++ {
		result := <-resultChannel

		if result.err != nil {
			rs.logger.Error("Provider error",
				"provider", result.provider,
				"error", result.err)
			continue
		}

		aggregatedRoutes = append(aggregatedRoutes, result.routes...)
		providersUsed = append(providersUsed, result.provider)
	}

	close(resultChannel)

	// Remove duplicates and sort
	uniqueRoutes := rs.removeDuplicates(aggregatedRoutes)
	sortedRoutes := rs.sortRoutes(uniqueRoutes)

	// Create response
	response := &models.RoutesResponse{
		Data: sortedRoutes,
		Metadata: models.ResponseMetadata{
			TotalCount:    len(sortedRoutes),
			ProvidersUsed: providersUsed,
			CacheHit:      false,
			Timestamp:     time.Now(),
		},
	}

	// Cache the response
	rs.cache.Set(cacheKey, response, 5*time.Minute)

	return response, nil
}

// GetHealth returns the health status of all providers
func (rs *RouteServiceImpl) GetHealth(ctx context.Context) (*models.HealthResponse, error) {
	rs.logger.Debug("Checking health status")

	// Get provider manager interface
	var providerStatuses map[string]string
	if pm, ok := rs.providerManager.(*providers.ProviderManager); ok {
		providerStatuses = pm.GetProviderStatuses(ctx)
	} else {
		// Fallback to interface methods
		providerStatuses = make(map[string]string)
		for _, provider := range rs.providerManager.GetAllProviders() {
			if err := provider.IsHealthy(ctx); err != nil {
				providerStatuses[provider.GetName()] = "unhealthy"
			} else {
				providerStatuses[provider.GetName()] = "healthy"
			}
		}
	}

	// Determine overall status
	overallStatus := "healthy"
	for _, status := range providerStatuses {
		if status == "unhealthy" {
			overallStatus = "degraded"
			break
		}
	}

	// If no providers are healthy, status is unhealthy
	hasHealthyProvider := false
	for _, status := range providerStatuses {
		if status == "healthy" {
			hasHealthyProvider = true
			break
		}
	}

	if !hasHealthyProvider {
		overallStatus = "unhealthy"
	}

	return &models.HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Providers: providerStatuses,
	}, nil
}

// removeDuplicates removes duplicate routes based on key fields
func (rs *RouteServiceImpl) removeDuplicates(routes []models.FlightRoute) []models.FlightRoute {
	seen := make(map[string]bool)
	uniqueRoutes := make([]models.FlightRoute, 0, len(routes))

	for _, route := range routes {
		// Create a unique key based on route fields
		key := fmt.Sprintf("%s-%s-%s-%s-%d",
			route.Airline,
			route.SourceAirport,
			route.DestinationAirport,
			route.CodeShare,
			route.Stops)

		if !seen[key] {
			seen[key] = true
			uniqueRoutes = append(uniqueRoutes, route)
		}
	}

	rs.logger.Debug("Removed duplicates",
		"original_count", len(routes),
		"unique_count", len(uniqueRoutes))

	return uniqueRoutes
}

// sortRoutes sorts routes by airline, source, destination, and stops
func (rs *RouteServiceImpl) sortRoutes(routes []models.FlightRoute) []models.FlightRoute {
	sort.Slice(routes, func(i, j int) bool {
		// Sort by airline first
		if routes[i].Airline != routes[j].Airline {
			return routes[i].Airline < routes[j].Airline
		}

		// Then by source airport
		if routes[i].SourceAirport != routes[j].SourceAirport {
			return routes[i].SourceAirport < routes[j].SourceAirport
		}

		// Then by destination airport
		if routes[i].DestinationAirport != routes[j].DestinationAirport {
			return routes[i].DestinationAirport < routes[j].DestinationAirport
		}

		// Finally by stops (fewer stops first)
		return routes[i].Stops < routes[j].Stops
	})

	return routes
}
