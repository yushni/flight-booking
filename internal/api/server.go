package api

import (
	"context"
	"net/http"
	"time"

	"flight-booking/internal/api/gen"
	"flight-booking/internal/api/handlers"
	"flight-booking/internal/services/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func NewServer(
	routeHandlers *handlers.RouteHandler,
	logger logger.Logger,
	lc fx.Lifecycle,
) {
	allHandlers := struct {
		*handlers.RouteHandler
	}{
		RouteHandler: routeHandlers,
	}

	engine := gin.New()
	gen.RegisterHandlersWithOptions(engine, allHandlers, gen.GinServerOptions{
		Middlewares: []gen.MiddlewareFunc{
			RequestID(),
			CORSHandler(),
			Panic(logger),
			RequestLogger(logger),
		},
		ErrorHandler: ErrorHandler(logger),
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: engine.Handler(),
	}

	lc.Append(fx.StartHook(func(ctx context.Context) error {
		go func() {
			// service connections
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("listen: %s\n", err)
			}
		}()

		return nil
	}))

	lc.Append(fx.StopHook(func(ctx context.Context) error {
		logger.Info("shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown failed", err)
		}

		return nil
	}))
}
