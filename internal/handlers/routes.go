package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"flight-booking/internal/interfaces"
	"flight-booking/internal/models"
)

// RouteHandler handles HTTP requests for flight routes
type RouteHandler struct {
	routeService interfaces.RouteService
	validator    *validator.Validate
	logger       interfaces.Logger
}

// NewRouteHandler creates a new route handler
func NewRouteHandler(
	routeService interfaces.RouteService,
	validator *validator.Validate,
	logger interfaces.Logger,
) *RouteHandler {
	return &RouteHandler{
		routeService: routeService,
		validator:    validator,
		logger:       logger.With("component", "route_handler"),
	}
}

// GetRoutes handles GET /routes requests
// @Summary Get flight routes
// @Description Retrieve aggregated flight route information from multiple providers
// @Tags routes
// @Accept json
// @Produce json
// @Param airline query string false "Filter by airline code"
// @Param sourceAirport query string false "Filter by source airport code"
// @Param destinationAirport query string false "Filter by destination airport code"
// @Param maxStops query int false "Maximum number of stops"
// @Success 200 {object} models.RoutesResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /routes [get]
func (h *RouteHandler) GetRoutes(c *gin.Context) {
	var filters models.RouteFilters

	// Bind query parameters
	if err := c.ShouldBindQuery(&filters); err != nil {
		h.logger.Error("Failed to bind query parameters", "error", err)
		h.sendErrorResponse(c, http.StatusBadRequest, "INVALID_QUERY_PARAMS", "Invalid query parameters")
		return
	}

	// Validate filters
	if err := h.validator.Struct(filters); err != nil {
		h.logger.Error("Validation failed", "error", err)
		h.sendErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed: "+err.Error())
		return
	}

	// Log the request
	h.logger.Info("Processing routes request",
		"filters", filters,
		"client_ip", c.ClientIP(),
		"user_agent", c.GetHeader("User-Agent"))

	// Call the service
	response, err := h.routeService.GetRoutes(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error("Service error", "error", err)
		h.sendErrorResponse(c, http.StatusInternalServerError, "SERVICE_ERROR", "Failed to fetch routes")
		return
	}

	// Log success
	h.logger.Info("Successfully returned routes",
		"total_routes", response.Metadata.TotalCount,
		"providers_used", len(response.Metadata.ProvidersUsed),
		"cache_hit", response.Metadata.CacheHit)

	c.JSON(http.StatusOK, response)
}

// GetHealth handles GET /health requests
// @Summary Health check
// @Description Check the health status of the API and all providers
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /health [get]
func (h *RouteHandler) GetHealth(c *gin.Context) {
	h.logger.Debug("Processing health check request")

	// Call the service
	response, err := h.routeService.GetHealth(c.Request.Context())
	if err != nil {
		h.logger.Error("Health check failed", "error", err)
		h.sendErrorResponse(c, http.StatusInternalServerError, "HEALTH_CHECK_ERROR", "Health check failed")
		return
	}

	// Determine HTTP status based on health status
	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	h.logger.Debug("Health check completed",
		"status", response.Status,
		"providers", response.Providers)

	c.JSON(statusCode, response)
}

// sendErrorResponse sends a standardized error response
func (h *RouteHandler) sendErrorResponse(c *gin.Context, statusCode int, errorCode, message string) {
	errorResponse := models.ErrorResponse{
		Error:     message,
		Code:      errorCode,
		Timestamp: time.Now(),
	}

	c.JSON(statusCode, errorResponse)
}

// RegisterRoutes registers all route handlers
func (h *RouteHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.GET("/routes", h.GetRoutes)
		v1.GET("/health", h.GetHealth)
	}

	// Also register health at root level for load balancer health checks
	router.GET("/health", h.GetHealth)
}

// Middleware functions

// RequestLogger logs HTTP requests
func RequestLogger(logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log after request
		duration := time.Since(start)
		logger.Info("HTTP request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"status", c.Writer.Status(),
			"duration", duration,
			"client_ip", c.ClientIP(),
			"user_agent", c.GetHeader("User-Agent"),
			"request_id", c.GetHeader("X-Request-ID"),
		)
	}
}

// ErrorHandler handles panics and errors
func ErrorHandler(logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered", "error", err)

				errorResponse := models.ErrorResponse{
					Error:     "Internal server error",
					Code:      "INTERNAL_ERROR",
					Timestamp: time.Now(),
				}

				c.JSON(http.StatusInternalServerError, errorResponse)
				c.Abort()
			}
		}()

		c.Next()
	}
}

// CORSHandler handles CORS headers
func CORSHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
