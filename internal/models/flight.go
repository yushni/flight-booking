package models

type RouteFilters struct {
	Airline            string
	SourceAirport      string
	DestinationAirport string
	MaxStops           *int
	Limit              int
	Offset             int
}

type Route struct {
	Airline            string  `json:"airline"`
	SourceAirport      string  `json:"sourceAirport"`
	DestinationAirport string  `json:"destinationAirport"`
	CodeShare          string  `json:"codeShare"`
	Stops              int     `json:"stops"`
	Equipment          *string `json:"equipment,omitempty"`
	Provider           string  `json:"provider"`
}
