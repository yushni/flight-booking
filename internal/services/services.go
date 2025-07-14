package services

import (
	"flight-booking/internal/services/cache"
	"flight-booking/internal/services/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			cache.New,
			logger.New,
		),
	)
}
