package config

import (
	"time"

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
	Server    ServerConfig
	Log       LogConfig
	Providers ProvidersConfig
}

type ProvidersConfig struct {
	Provider1BaseURL  string        `env:"PROVIDER1_BASE_URL"  envDefault:"https://4r5rvu2fcydfzr5gymlhcsnfem0lyxoe.lambda-url.eu-central-1.on.aws/provider/flights1"`
	Provider1Timeout  time.Duration `env:"PROVIDER1_TIMEOUT"   envDefault:"30s"`
	Provider1CacheTTL time.Duration `env:"PROVIDER1_CACHE_TTL" envDefault:"60s"`

	Provider2BaseURL  string        `env:"PROVIDER2_BASE_URL"  envDefault:"https://4r5rvu2fcydfzr5gymlhcsnfem0lyxoe.lambda-url.eu-central-1.on.aws/provider/flights2"`
	Provider2Timeout  time.Duration `env:"PROVIDER2_TIMEOUT"   envDefault:"30s"`
	Provider2CacheTTL time.Duration `env:"PROVIDER2_CACHE_TTL" envDefault:"60s"`
}

type ServerConfig struct {
	Port string `env:"SERVER_PORT" envDefault:":8080"`
	Host string `env:"SERVER_HOST" envDefault:"localhost"`
}

type LogConfig struct {
	Level  string `env:"LOG_LEVEL"  envDefault:"debug"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}
