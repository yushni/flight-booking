package providers

import (
	"context"
	"sync"

	"flight-booking/internal/interfaces"
	"flight-booking/internal/models"
)

// ProviderManager manages multiple route providers
type ProviderManager struct {
	providers map[string]interfaces.RouteProvider
	logger    interfaces.Logger
	mu        sync.RWMutex
}

// NewProviderManager creates a new provider manager
func NewProviderManager(providers []interfaces.RouteProvider, logger interfaces.Logger) interfaces.ProviderManager {
	providerMap := make(map[string]interfaces.RouteProvider)

	for _, provider := range providers {
		providerMap[provider.GetName()] = provider
	}

	return &ProviderManager{
		providers: providerMap,
		logger:    logger.With("component", "provider_manager"),
	}
}

// GetAllProviders returns all registered providers
func (pm *ProviderManager) GetAllProviders() []interfaces.RouteProvider {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	providers := make([]interfaces.RouteProvider, 0, len(pm.providers))
	for _, provider := range pm.providers {
		providers = append(providers, provider)
	}

	return providers
}

// GetEnabledProviders returns only enabled providers
func (pm *ProviderManager) GetEnabledProviders() []interfaces.RouteProvider {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var enabledProviders []interfaces.RouteProvider

	for _, provider := range pm.providers {
		// Check if provider is healthy (which includes enabled check)
		if err := provider.IsHealthy(context.Background()); err == nil {
			enabledProviders = append(enabledProviders, provider)
		}
	}

	return enabledProviders
}

// GetProvider returns a specific provider by name
func (pm *ProviderManager) GetProvider(name string) (interfaces.RouteProvider, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	provider, exists := pm.providers[name]
	return provider, exists
}

// AddProvider adds a new provider
func (pm *ProviderManager) AddProvider(provider interfaces.RouteProvider) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.providers[provider.GetName()] = provider
	pm.logger.Info("Provider added", "name", provider.GetName())
}

// RemoveProvider removes a provider
func (pm *ProviderManager) RemoveProvider(name string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delete(pm.providers, name)
	pm.logger.Info("Provider removed", "name", name)
}

// GetProviderStatuses returns the health status of all providers
func (pm *ProviderManager) GetProviderStatuses(ctx context.Context) map[string]string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	statuses := make(map[string]string)

	for name, provider := range pm.providers {
		if err := provider.IsHealthy(ctx); err != nil {
			statuses[name] = "unhealthy"
		} else {
			statuses[name] = "healthy"
		}
	}

	return statuses
}

// GetRoutesFromAllProviders fetches routes from all enabled providers concurrently
func (pm *ProviderManager) GetRoutesFromAllProviders(ctx context.Context, filters models.RouteFilters) []models.ProviderResponse {
	enabledProviders := pm.GetEnabledProviders()

	if len(enabledProviders) == 0 {
		pm.logger.Warn("No enabled providers found")
		return []models.ProviderResponse{}
	}

	// Create channels for concurrent execution
	responseChannel := make(chan models.ProviderResponse, len(enabledProviders))

	// Launch goroutines for each provider
	for _, provider := range enabledProviders {
		go func(p interfaces.RouteProvider) {
			routes, err := p.GetRoutes(ctx, filters)
			responseChannel <- models.ProviderResponse{
				Routes:   routes,
				Provider: p.GetName(),
				Error:    err,
			}
		}(provider)
	}

	// Collect responses
	responses := make([]models.ProviderResponse, 0, len(enabledProviders))
	for i := 0; i < len(enabledProviders); i++ {
		response := <-responseChannel
		responses = append(responses, response)

		if response.Error != nil {
			pm.logger.Error("Provider returned error",
				"provider", response.Provider,
				"error", response.Error)
		} else {
			pm.logger.Debug("Provider returned routes",
				"provider", response.Provider,
				"count", len(response.Routes))
		}
	}

	close(responseChannel)
	return responses
}
