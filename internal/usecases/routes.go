package usecases

import (
	"context"

	"flight-booking/internal/models"
	"flight-booking/internal/services/providers"
)

type Routes interface {
	GetRoutes(ctx context.Context, filters models.RouteFilters) (*models.RoutesResponse, error)
}

type routes struct {
	provider providers.Provider
}

func NewRoutes(provider providers.Provider) Routes {
	return &routes{
		provider: provider,
	}
}

func (r *routes) GetRoutes(ctx context.Context, filters models.RouteFilters) (*models.RoutesResponse, error) {
	routes, err := r.provider.GetRoutes(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &models.RoutesResponse{
		Data: routes,
	}, nil
}
