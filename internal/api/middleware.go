package api

import (
	"net/http"
	"time"

	"flight-booking/internal/api/gen"
	"flight-booking/internal/services/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		logger.Context(c.Request.Context()).Info("HTTP request",
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

func Panic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Context(c.Request.Context()).Error("Panic recovered", "error", err)

				errorResponse := gen.ErrorResponse{
					Error:     "Internal server error",
					Code:      http.StatusInternalServerError,
					Timestamp: time.Now(),
				}

				c.JSON(http.StatusInternalServerError, errorResponse)
				c.Abort()
			}
		}()

		c.Next()
	}
}

func ErrorHandler() func(*gin.Context, error, int) {
	return func(c *gin.Context, _ error, _ int) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Context(c.Request.Context()).Error("Error in API handler", "error", err.Err)
			}

			errorResponse := gen.ErrorResponse{
				Error:     "Internal server error",
				Code:      http.StatusInternalServerError,
				Timestamp: time.Now(),
			}

			c.JSON(http.StatusInternalServerError, errorResponse)
			c.Abort()

			return
		}
	}
}

func RequestID() gin.HandlerFunc {
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

func ContextLogger(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		l := logger.With("request_id", c.GetString("request_id"))

		ctx = l.SetIntoContext(ctx)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
