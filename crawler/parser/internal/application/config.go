package application

import "github.com/vitalyshatskikh/go-lib/config"

type Config struct {
	config.Config

	TemporalHost      string `env:"TEMPORAL_HOST" env-default:"localhost:7233"`
	TemporalNamespace string `env:"TEMPORAL_NAMESPACE" env-default:"crawler"`
}
