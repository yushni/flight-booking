package providers

import (
	"context"
	"errors"
	"fmt"

	"flight-booking/internal/config"
	"flight-booking/internal/models"
	"flight-booking/internal/services/cache"
	"flight-booking/internal/services/logger"
	"resty.dev/v3"
)

const (
	defaultRetryCount = 3
)

type Provider interface {
	GetRoutes(ctx context.Context, filters models.RouteFilters) ([]models.Route, error)
}

type provider struct {
	config          config.Config
	cache           cache.Cache
	provider1Client *resty.Client
	provider2Client *resty.Client
}

func New(config config.Config, cache cache.Cache) Provider {
	p := provider{
		config: config,
		cache:  cache,
	}

	p.provider1Client = resty.New().
		SetBaseURL(config.Providers.Provider1BaseURL).
		SetTimeout(config.Providers.Provider1Timeout).
		SetRetryCount(defaultRetryCount).
		SetCircuitBreaker(resty.NewCircuitBreaker())

	p.provider2Client = resty.New().
		SetBaseURL(config.Providers.Provider2BaseURL).
		SetTimeout(config.Providers.Provider2Timeout).
		SetRetryCount(defaultRetryCount).
		SetCircuitBreaker(resty.NewCircuitBreaker())

	return p
}

func (p provider) GetRoutes(ctx context.Context, filters models.RouteFilters) ([]models.Route, error) {
	var routes []models.Route
	routes1, err := p.routesFromProvider1(ctx)
	if err != nil {
		logger.Context(ctx).Error("error fetching routes from provider1", "error", err)
	}

	routes = append(routes, routes1...)
	routes2, err := p.routesFromProvider2(ctx)
	if err != nil {
		logger.Context(ctx).Error("error fetching routes from provider2", "error", err)
	}

	routes = append(routes, routes2...)
	return p.ApplyFilters(filters, routes), nil
}

func (p provider) routesFromProvider1(ctx context.Context) ([]models.Route, error) {
	data, err := p.cache.GetOrLoad("provider1_routes", p.config.Providers.Provider1CacheTTL, func() (interface{}, error) {
		var res []models.Route

		resp, err := p.provider1Client.R().
			SetContext(ctx).
			SetResult(&res).
			Get("")
		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("provider1 request failed: %s", resp.String())
		}

		return res, nil
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching routes from cache or provider1: %w", err)
	}

	if routes, ok := data.([]models.Route); ok {
		return routes, nil
	}

	return nil, errors.New("unexpected data type from cache for provider1 routes")
}

func (p provider) routesFromProvider2(ctx context.Context) ([]models.Route, error) {
	data, err := p.cache.GetOrLoad("provider2_routes", p.config.Providers.Provider2CacheTTL, func() (interface{}, error) {
		var res []models.Route

		resp, err := p.provider2Client.R().
			SetContext(ctx).
			SetResult(&res).
			Get("")
		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("provider2 request failed: %s", resp.String())
		}

		return res, nil
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching routes from cache or provider2: %w", err)
	}

	if routes, ok := data.([]models.Route); ok {
		return routes, nil
	}

	return nil, errors.New("unexpected data type from cache for provider2 routes")
}

func (p provider) ApplyFilters(filters models.RouteFilters, routes []models.Route) []models.Route {
	if len(routes) == 0 {
		return routes
	}
	if filters.Limit == 0 {
		filters.Limit = 100
	}

	filtered := make([]models.Route, 0, len(routes))

	skipped := 0
	for _, route := range routes {
		if filters.Airline != "" && route.Airline != filters.Airline {
			continue
		}

		if filters.SourceAirport != "" && route.SourceAirport != filters.SourceAirport {
			continue
		}

		if filters.DestinationAirport != "" && route.DestinationAirport != filters.DestinationAirport {
			continue
		}

		if filters.MaxStops != nil && route.Stops > *filters.MaxStops {
			continue
		}

		if filters.Limit > 0 && len(filtered) >= filters.Limit {
			break
		}

		if filters.Offset > 0 && skipped < filters.Offset {
			skipped++
			continue
		}

		filtered = append(filtered, route)
	}

	return filtered
}
