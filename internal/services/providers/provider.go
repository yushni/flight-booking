package providers

import (
	"context"
	"fmt"

	"flight-booking/internal/config"
	"flight-booking/internal/models"
	"flight-booking/internal/services/cache"
	"resty.dev/v3"
)

const (
	defaultRetryCount = 3
)

type Provider interface {
	GetRoutes(ctx context.Context, filters models.RouteFilters) ([]models.FlightRoute, error)
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

func (p provider) GetRoutes(ctx context.Context, filters models.RouteFilters) ([]models.FlightRoute, error) {
	var routes []models.FlightRoute
	var err error

	routes1, err := p.routesFromProvider1(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("error fetching routes from provider1: %w", err)
	}
	routes = append(routes, routes1...)

	routes2, err := p.routesFromProvider2(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("error fetching routes from provider2: %w", err)
	}
	routes = append(routes, routes2...)

	return routes, nil
}

func (p provider) routesFromProvider1(ctx context.Context, filters models.RouteFilters) ([]models.FlightRoute, error) {
	data, err := p.cache.GetOrLoad("provider1_routes", p.config.Providers.Provider1CacheTTL, func() (interface{}, error) {
		var res []models.FlightRoute

		resp, err := p.provider1Client.R().
			SetContext(ctx).
			//SetQueryParamsFromValues(filters.ToQueryParams()).
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
	if routes, ok := data.([]models.FlightRoute); ok {
		return routes, nil
	}

	return nil, fmt.Errorf("unexpected data type from cache for provider1 routes")
}

func (p provider) routesFromProvider2(ctx context.Context, filters models.RouteFilters) ([]models.FlightRoute, error) {
	data, err := p.cache.GetOrLoad("provider2_routes", p.config.Providers.Provider2CacheTTL, func() (interface{}, error) {
		var res []models.FlightRoute

		resp, err := p.provider2Client.R().
			SetContext(ctx).
			//SetQueryParamsFromValues(filters.ToQueryParams()).
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
	if routes, ok := data.([]models.FlightRoute); ok {
		return routes, nil
	}

	return nil, fmt.Errorf("unexpected data type from cache for provider2 routes")
}
