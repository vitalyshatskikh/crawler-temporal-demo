package advertswf

import (
	"time"
)

var DefaultSurferConfig = SurferConfig{
	ProcessBranchTimeout:         15 * time.Minute,
	ProcessSearchPageTimeout:     5 * time.Minute,
	ProcessAdvertTimeout:         5 * time.Minute,
	DownloadSearchPageTimeout:    4 * time.Minute,
	DownloadAdvertContentTimeout: 4 * time.Minute,
	RepoRequestTimeout:           15 * time.Second,
	RepoRequestRetry:             RetryConfig{MaxAttempts: 3},
	ParseSearchPageTimeout:       30 * time.Second,
	ParseSearchPageRetry:         RetryConfig{MaxAttempts: 3},
	ParseAdvertContentTimeout:    30 * time.Second,
	ParseAdvertContentRetry:      RetryConfig{MaxAttempts: 3},
}

type SurferConfig struct {
	ProcessBranchTimeout     time.Duration `validate:"min=1000000000"`
	ProcessSearchPageTimeout time.Duration `validate:"min=1000000000"`
	ProcessAdvertTimeout     time.Duration `validate:"min=1000000000"`

	DownloadSearchPageTimeout    time.Duration `validate:"min=1000000000"`
	DownloadAdvertContentTimeout time.Duration `validate:"min=1000000000"`

	RepoRequestTimeout time.Duration `validate:"min=1000000000"`
	RepoRequestRetry   RetryConfig

	ParseSearchPageTimeout time.Duration `validate:"min=1000000000"`
	ParseSearchPageRetry   RetryConfig

	ParseAdvertContentTimeout time.Duration `validate:"min=1000000000"`
	ParseAdvertContentRetry   RetryConfig
}

type RetryConfig struct {
	MaxAttempts        int32 `validate:"min=0"`
	InitInterval       time.Duration
	MaxInterval        time.Duration
	BackoffCoefficient float64 `validate:"min=0.0"`
}
