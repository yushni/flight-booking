// Package gen provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package gen

import (
	"time"
)

// Defines values for FlightRouteCodeShare.
const (
	N FlightRouteCodeShare = "N"
	Y FlightRouteCodeShare = "Y"
)

// ErrorResponse defines model for ErrorResponse.
type ErrorResponse struct {
	// Code HTTP status code
	Code int `json:"code"`

	// Error Error message
	Error string `json:"error"`

	// Timestamp Error timestamp
	Timestamp time.Time `json:"timestamp"`
}

// FlightRoute defines model for FlightRoute.
type FlightRoute struct {
	// Airline Airline code (IATA 2-letter code)
	Airline string `json:"airline"`

	// CodeShare Code share information
	CodeShare FlightRouteCodeShare `json:"codeShare"`

	// DestinationAirport Destination airport code (IATA 3-letter code)
	DestinationAirport string `json:"destinationAirport"`

	// Equipment Equipment type (optional)
	Equipment *string `json:"equipment"`

	// Provider Data provider source
	Provider *string `json:"provider,omitempty"`

	// SourceAirport Source airport code (IATA 3-letter code)
	SourceAirport string `json:"sourceAirport"`

	// Stops Number of stops
	Stops int `json:"stops"`
}

// FlightRouteCodeShare Code share information
type FlightRouteCodeShare string

// RoutesResponse defines model for RoutesResponse.
type RoutesResponse struct {
	// Data Array of flight routes
	Data []FlightRoute `json:"data"`
}

// GetRoutesParams defines parameters for GetRoutes.
type GetRoutesParams struct {
	// Airline Filter by airline code
	Airline *string `form:"airline,omitempty" json:"airline,omitempty"`

	// SourceAirport Filter by source airport code
	SourceAirport *string `form:"sourceAirport,omitempty" json:"sourceAirport,omitempty"`

	// DestinationAirport Filter by destination airport code
	DestinationAirport *string `form:"destinationAirport,omitempty" json:"destinationAirport,omitempty"`

	// MaxStops Maximum number of stops
	MaxStops *int `form:"maxStops,omitempty" json:"maxStops,omitempty"`

	// Limit Maximum number of routes to return
	Limit *int `form:"limit,omitempty" json:"limit,omitempty"`

	// Offset Offset for pagination
	Offset *int `form:"offset,omitempty" json:"offset,omitempty"`
}
