package main

import (
	"time"

	"flight-booking/config"
	"flight-booking/internal/api"
	"flight-booking/internal/services"
	"flight-booking/internal/usecases"
	"go.uber.org/fx"
)

func main() {
	conf := config.Config{}

	app := fx.New(
		fx.Supply(conf),
		fx.StopTimeout(time.Second*20),
		api.Module(),
		services.Module(),
		usecases.Module(),
	)

	app.Run()
}
