package cleaner

import "time"

type CleanConfig struct {
	CleanupInterval time.Duration `env:"CLEANUP_INTERVAL" env-default:"10m"`
	CleanupDuration time.Duration `env:"CLEANUP_DURATION" env-default:"24h"`
}
