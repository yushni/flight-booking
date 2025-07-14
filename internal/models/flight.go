package models

import (
	"fmt"
	"time"
)

// FlightRoute represents a flight route
type FlightRoute struct {
	Airline            string  `json:"airline" validate:"required"`
	SourceAirport      string  `json:"sourceAirport" validate:"required"`
	DestinationAirport string  `json:"destinationAirport" validate:"required"`
	CodeShare          string  `json:"codeShare" validate:"required"`
	Stops              int     `json:"stops" validate:"min=0"`
	Equipment          *string `json:"equipment,omitempty"`
	Provider           string  `json:"provider"`
}

// RoutesResponse represents the response for flight routes
type RoutesResponse struct {
	Data     []FlightRoute    `json:"data"`
	Metadata ResponseMetadata `json:"metadata"`
}

// ResponseMetadata represents metadata for the response
type ResponseMetadata struct {
	TotalCount    int       `json:"totalCount"`
	ProvidersUsed []string  `json:"providersUsed"`
	CacheHit      bool      `json:"cacheHit"`
	Timestamp     time.Time `json:"timestamp"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Providers map[string]string `json:"providers"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string    `json:"error"`
	Code      string    `json:"code"`
	Timestamp time.Time `json:"timestamp"`
}

// RouteFilters represents filters for route queries
type RouteFilters struct {
	Airline            string `form:"airline"`
	SourceAirport      string `form:"sourceAirport"`
	DestinationAirport string `form:"destinationAirport"`
	MaxStops           *int   `form:"maxStops" validate:"omitempty,min=0"`
}

// ProviderResponse represents the response from a provider
type ProviderResponse struct {
	Routes   []FlightRoute `json:"routes"`
	Provider string        `json:"provider"`
	Error    error         `json:"error,omitempty"`
}

// CacheKey represents a cache key for routes
type CacheKey struct {
	Filters RouteFilters `json:"filters"`
}

// String returns the string representation of the cache key
func (ck CacheKey) String() string {
	return fmt.Sprintf("routes:%s:%s:%s:%d",
		ck.Filters.Airline,
		ck.Filters.SourceAirport,
		ck.Filters.DestinationAirport,
		getIntValue(ck.Filters.MaxStops))
}

// getIntValue returns the value of an int pointer or -1 if nil
func getIntValue(ptr *int) int {
	if ptr == nil {
		return -1
	}
	return *ptr
}

// ApplyFilters applies filters to a slice of flight routes
func (filters RouteFilters) ApplyFilters(routes []FlightRoute) []FlightRoute {
	if len(routes) == 0 {
		return routes
	}

	filtered := make([]FlightRoute, 0, len(routes))

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

		filtered = append(filtered, route)
	}

	return filtered
}

// Validate validates the flight route
func (fr *FlightRoute) Validate() error {
	if fr.Airline == "" {
		return fmt.Errorf("airline is required")
	}

	if fr.SourceAirport == "" {
		return fmt.Errorf("source airport is required")
	}

	if fr.DestinationAirport == "" {
		return fmt.Errorf("destination airport is required")
	}

	if fr.CodeShare == "" {
		return fmt.Errorf("code share is required")
	}

	if fr.Stops < 0 {
		return fmt.Errorf("stops must be non-negative")
	}

	return nil
}
