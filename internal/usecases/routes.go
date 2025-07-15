package usecases

import (
	"context"
	"fmt"

	"flight-booking/internal/models"
	"flight-booking/internal/services/providers"
)

type Routes interface {
	GetRoutes(ctx context.Context, filters models.RouteFilters) ([]models.Route, error)
}

type routes struct {
	provider providers.Provider
}

func NewRoutes(provider providers.Provider) Routes {
	return &routes{
		provider: provider,
	}
}

func (r *routes) GetRoutes(ctx context.Context, filters models.RouteFilters) ([]models.Route, error) {
	routes, err := r.provider.GetRoutes(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes from provider: %w", err)
	}

	return routes, nil
}
