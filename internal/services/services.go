package services

import (
	"flight-booking/internal/services/cache"
	"flight-booking/internal/services/logger"
	"flight-booking/internal/services/providers"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			cache.New,
			logger.New,
			providers.New,
		),
	)
}
