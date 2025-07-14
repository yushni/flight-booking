package config

import (
	"github.com/caarlos0/env/v11"
)

func New() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

type Config struct {
	Server ServerConfig
	Log    LogConfig
}

type ServerConfig struct {
	Port string `env:"SERVER_PORT" envDefault:":8080"`
	Host string `env:"SERVER_HOST" envDefault:"localhost"`
}

type LogConfig struct {
	Level  string `env:"LOG_LEVEL"  envDefault:"debug"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}
