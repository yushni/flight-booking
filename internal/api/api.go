package api

import (
	"flight-booking/internal/api/handlers"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			handlers.NewRouteHandler,
			handlers.NewHealthHandler,
		),
		fx.Invoke(NewServer),
	)
}
