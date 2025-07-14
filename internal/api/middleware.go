package api

import (
	"net/http"
	"time"

	"flight-booking/internal/api/gen"
	"flight-booking/internal/services/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestLogger logs HTTP requests
func RequestLogger(logger logger.Logger) gen.MiddlewareFunc {
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
			"request_id", c.GetString("request_id"),
		)
	}
}

// Panic handles panics and errors
func Panic(logger logger.Logger) gen.MiddlewareFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered", "error", err)

				errorResponse := gen.ErrorResponse{
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

// ErrorHandler handles errors and sends appropriate responses
func ErrorHandler(logger logger.Logger) func(*gin.Context, error, int) {
	return func(c *gin.Context, err error, i int) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("Error in API handler", "error", err.Err)
			}

			errorResponse := gen.ErrorResponse{
				Error:     "Internal server error",
				Code:      "INTERNAL_ERROR",
				Timestamp: time.Now(),
			}

			c.JSON(http.StatusInternalServerError, errorResponse)
			c.Abort()
			return
		}
	}

}

// CORSHandler handles CORS headers
func CORSHandler() gen.MiddlewareFunc {
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
func RequestID() gen.MiddlewareFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}
