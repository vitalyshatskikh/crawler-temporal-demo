//go:generate go tool mockery
package domain

import (
	"context"
)

type ConfigRepository interface {
	GetConfig(ctx context.Context, sourceID SourceID, docType DocumentType) (ParsingConfig, error)
}
