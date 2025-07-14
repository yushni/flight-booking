package handlers

import (
	"net/http"
	"time"

	"flight-booking/internal/api/gen"
	"flight-booking/internal/services/logger"
	"github.com/gin-gonic/gin"

	"flight-booking/internal/models"
	"flight-booking/internal/usecases"
)

type RouteHandler struct {
	routeService usecases.Routes
	logger       logger.Logger
}

// NewRouteHandler creates a new route handler
func NewRouteHandler(routeService usecases.Routes, logger logger.Logger) *RouteHandler {
	return &RouteHandler{
		routeService: routeService,
		logger:       logger.With("component", "route_handler"),
	}
}

// GetRoutes implements the GetRoutes method from ServerInterface
func (h *RouteHandler) GetRoutes(c *gin.Context, params gen.GetRoutesParams) {
	filters := h.convertParamsToFilters(params)

	response, err := h.routeService.GetRoutes(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error("Service error", "error", err)
		h.sendErrorResponse(c, http.StatusInternalServerError, "SERVICE_ERROR", "Failed to fetch routes")
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

func (h *RouteHandler) convertToAPIResponse(response *models.RoutesResponse) *gen.RoutesResponse {
	if response == nil {
		return &gen.RoutesResponse{
			Data: []gen.FlightRoute{},
			Metadata: gen.ResponseMetadata{
				TotalCount:    0,
				ProvidersUsed: []string{},
				CacheHit:      false,
				Timestamp:     time.Now(),
			},
		}
	}

	apiRoutes := make([]gen.FlightRoute, len(response.Data))

	for i, route := range response.Data {
		apiRoute := gen.FlightRoute{
			Airline:            route.Airline,
			SourceAirport:      route.SourceAirport,
			DestinationAirport: route.DestinationAirport,
			CodeShare:          gen.FlightRouteCodeShare(route.CodeShare),
			Stops:              route.Stops,
		}

		if route.Equipment != nil {
			apiRoute.Equipment = route.Equipment
		}

		if route.Provider != "" {
			apiRoute.Provider = &route.Provider
		}

		apiRoutes[i] = apiRoute
	}

	return &gen.RoutesResponse{
		Data: apiRoutes,
		Metadata: gen.ResponseMetadata{
			TotalCount:    response.Metadata.TotalCount,
			ProvidersUsed: response.Metadata.ProvidersUsed,
			CacheHit:      response.Metadata.CacheHit,
			Timestamp:     response.Metadata.Timestamp,
		},
	}
}

func (h *RouteHandler) sendErrorResponse(c *gin.Context, statusCode int, errorCode, message string) {
	errorResponse := gen.ErrorResponse{
		Error:     message,
		Code:      errorCode,
		Timestamp: time.Now(),
	}

	c.JSON(statusCode, errorResponse)
}
