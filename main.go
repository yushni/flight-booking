package main

import (
	"time"

	"flight-booking/internal/api"
	"flight-booking/internal/config"
	"flight-booking/internal/services"
	"flight-booking/internal/usecases"
	"go.uber.org/fx"
)

const (
	MaxShutdownTime = 20 * time.Second
)

func main() {
	conf, err := config.New()
	if err != nil {
		panic(err)
	}

	app := fx.New(
		fx.Supply(conf),
		fx.StopTimeout(MaxShutdownTime),
		api.Module(),
		services.Module(),
		usecases.Module(),
	)

	app.Run()
}
