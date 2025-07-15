package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"flight-booking/internal/api/gen"
	"flight-booking/internal/api/handlers"
	"flight-booking/internal/services/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func NewServer(routeHandlers *handlers.RouteHandler, logger logger.Logger, lc fx.Lifecycle) {
	allHandlers := struct {
		*handlers.RouteHandler
	}{
		RouteHandler: routeHandlers,
	}

	engine := gin.New()
	gen.RegisterHandlersWithOptions(engine, allHandlers, gen.GinServerOptions{
		Middlewares: []gen.MiddlewareFunc{
			RequestID(),
			ContextLogger(logger),
			Panic(),
			RequestLogger(),
		},
		ErrorHandler: ErrorHandler(),
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           engine.Handler(),
		ReadHeaderTimeout: 60 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
	}

	lc.Append(fx.StartHook(func(_ context.Context) error {
		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("listen: %s\n", err)
			}
		}()

		return nil
	}))

	lc.Append(fx.StopHook(func(ctx context.Context) error {
		logger.Info("shutting down server...")

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("shutdown failed", err)
		}

		return nil
	}))
}
