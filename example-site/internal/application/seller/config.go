package seller

import "time"

type Config struct {
	Region          string        `env:"REGION"`
	CreateRPS       int           `env:"CREATE_RPS" env-default:"3"`
	DeleteAge       time.Duration `env:"DELETE_AGE" env-default:"1h"`
	DeleteJitter    time.Duration `env:"DELETE_JITTER" env-default:"15m"`
	DeleteInterval  time.Duration `env:"DELETE_INTERVAL" env-default:"5m"`
	DeleteBatchSize int           `env:"DELETE_BATCH_SIZE" env-default:"1000"`
}
