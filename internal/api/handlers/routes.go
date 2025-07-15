package handlers

import (
	"net/http"

	"flight-booking/internal/api/gen"
	"flight-booking/internal/models"
	"flight-booking/internal/services/logger"
	"flight-booking/internal/usecases"
	"github.com/gin-gonic/gin"
)

type RouteHandler struct {
	routeService usecases.Routes
	logger       logger.Logger
}

// NewRouteHandler creates a new route handler.
func NewRouteHandler(routeService usecases.Routes, logger logger.Logger) *RouteHandler {
	return &RouteHandler{
		routeService: routeService,
		logger:       logger.With("component", "route_handler"),
	}
}

// GetRoutes implements the GetRoutes method from ServerInterface.
func (h *RouteHandler) GetRoutes(c *gin.Context, params gen.GetRoutesParams) {
	ctx := c.Request.Context()
	filters := h.convertParamsToFilters(params)

	response, err := h.routeService.GetRoutes(ctx, filters)
	if err != nil {
		_ = c.Error(err)

		return
	}

	apiResponse := h.convertToAPIResponse(response)
	c.JSON(http.StatusOK, apiResponse)
}

func (h *RouteHandler) convertParamsToFilters(params gen.GetRoutesParams) models.RouteFilters {
	filters := models.RouteFilters{}

	if params.Airline != nil {
		filters.Airline = *params.Airline
	}

	if params.SourceAirport != nil {
		filters.SourceAirport = *params.SourceAirport
	}

	if params.DestinationAirport != nil {
		filters.DestinationAirport = *params.DestinationAirport
	}

	if params.MaxStops != nil {
		filters.MaxStops = params.MaxStops
	}

	return filters
}

func (h *RouteHandler) convertToAPIResponse(routes []models.Route) *gen.RoutesResponse {
	apiRoutes := make([]gen.FlightRoute, len(routes))

	for i, route := range routes {
		apiRoute := gen.FlightRoute{
			Airline:            route.Airline,
			SourceAirport:      route.SourceAirport,
			DestinationAirport: route.DestinationAirport,
			CodeShare:          gen.FlightRouteCodeShare(route.CodeShare),
			Stops:              route.Stops,
			Equipment:          route.Equipment,
			Provider:           &route.Provider,
		}

		apiRoutes[i] = apiRoute
	}

	return &gen.RoutesResponse{
		Data: apiRoutes,
	}
}
