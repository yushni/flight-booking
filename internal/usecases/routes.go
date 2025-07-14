package usecases

import (
	"context"

	"flight-booking/internal/models"
)

type Routes interface {
	GetRoutes(ctx context.Context, filters models.RouteFilters) (*models.RoutesResponse, error)
}

type routes struct{}

func NewRoutes() Routes {
	return &routes{}
}

func (r *routes) GetRoutes(ctx context.Context, filters models.RouteFilters) (*models.RoutesResponse, error) {
	return nil, nil
}
