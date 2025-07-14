package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	"flight-booking/internal/config"
	"flight-booking/internal/handlers"
	"flight-booking/internal/interfaces"
	"flight-booking/internal/providers"
	"flight-booking/internal/services"
	"flight-booking/internal/usecases"
)

// Application represents the main application
type Application struct {
	config  *config.Config
	server  *http.Server
	logger  interfaces.Logger
	cleanup func()
}

// NewApplication creates a new application instance
func NewApplication() (*Application, error) {
	app := &Application{}

	// Create FX application
	fxApp := fx.New(
		// Provide dependencies
		fx.Provide(
			// Configuration
			config.Load,

			// Logger
			NewLogger,

			// Validator
			NewValidator,

			// Services
			NewCacheService,
			NewHTTPClient,

			// Providers
			NewProvider1,
			NewProvider2,
			NewProviderManager,

			// Use cases
			NewRouteService,

			// Handlers
			NewRouteHandler,

			// Server
			NewGinEngine,
			NewHTTPServer,
		),

		// Invoke lifecycle hooks
		fx.Invoke(RegisterRoutes),

		// Disable FX logs for cleaner output
		fx.WithLogger(func(logger interfaces.Logger) fx.Printer {
			return &fxLogger{logger: logger}
		}),
	)

	// Store FX app for cleanup
	app.cleanup = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		fxApp.Stop(ctx)
	}

	// Start the application
	if err := fxApp.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start application: %w", err)
	}

	return app, nil
}

// Run starts the application
func (a *Application) Run() error {
	if a.server == nil {
		return fmt.Errorf("server not initialized")
	}

	a.logger.Info("Starting HTTP server", "address", a.server.Addr)
	return a.server.ListenAndServe()
}

// Shutdown gracefully shuts down the application
func (a *Application) Shutdown(ctx context.Context) error {
	if a.cleanup != nil {
		a.cleanup()
	}

	if a.server != nil {
		return a.server.Shutdown(ctx)
	}

	return nil
}

// Dependency constructors

// NewLogger creates a new logger
func NewLogger(cfg *config.Config) (interfaces.Logger, error) {
	return services.NewZapLogger(cfg.Log)
}

// NewValidator creates a new validator
func NewValidator() *validator.Validate {
	return validator.New()
}

// NewCacheService creates a new cache service
func NewCacheService(cfg *config.Config) interfaces.CacheService {
	return services.NewMemoryCacheService(cfg.Cache)
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(cfg *config.Config) interfaces.HTTPClient {
	return services.NewHTTPClientService(cfg.Providers.Provider1.Timeout)
}

// NewProvider1 creates a new Provider1
func NewProvider1(
	cfg *config.Config,
	httpClient interfaces.HTTPClient,
	cache interfaces.CacheService,
	logger interfaces.Logger,
) interfaces.RouteProvider {
	return providers.NewProvider1(cfg.Providers.Provider1, httpClient, cache, logger)
}

// NewProvider2 creates a new Provider2
func NewProvider2(
	cfg *config.Config,
	httpClient interfaces.HTTPClient,
	cache interfaces.CacheService,
	logger interfaces.Logger,
) interfaces.RouteProvider {
	return providers.NewProvider2(cfg.Providers.Provider2, httpClient, cache, logger)
}

// NewProviderManager creates a new provider manager
func NewProviderManager(
	provider1 interfaces.RouteProvider,
	provider2 interfaces.RouteProvider,
	logger interfaces.Logger,
) interfaces.ProviderManager {
	providersList := []interfaces.RouteProvider{provider1, provider2}
	return providers.NewProviderManager(providersList, logger)
}

// NewRouteService creates a new route service
func NewRouteService(
	providerManager interfaces.ProviderManager,
	cache interfaces.CacheService,
	logger interfaces.Logger,
) interfaces.RouteService {
	return usecases.NewRouteService(providerManager, cache, logger)
}

// NewRouteHandler creates a new route handler
func NewRouteHandler(
	routeService interfaces.RouteService,
	validator *validator.Validate,
	logger interfaces.Logger,
) *handlers.RouteHandler {
	return handlers.NewRouteHandler(routeService, validator, logger)
}

// NewGinEngine creates a new Gin engine
func NewGinEngine(cfg *config.Config, logger interfaces.Logger) *gin.Engine {
	// Set Gin mode
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create engine
	engine := gin.New()

	// Add middleware
	engine.Use(
		handlers.RequestID(),
		handlers.RequestLogger(logger),
		handlers.ErrorHandler(logger),
		handlers.CORSHandler(),
		gin.Recovery(),
	)

	return engine
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(cfg *config.Config, engine *gin.Engine, logger interfaces.Logger) *http.Server {
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return server
}

// RegisterRoutes registers all application routes
func RegisterRoutes(handler *handlers.RouteHandler, engine *gin.Engine) {
	handler.RegisterRoutes(engine)
}

// fxLogger implements fx.Printer for custom logging
type fxLogger struct {
	logger interfaces.Logger
}

func (l *fxLogger) Printf(format string, v ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, v...))
}
