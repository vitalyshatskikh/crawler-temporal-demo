package adverts

import (
	"context"
	"time"
)

// SdocID is 'static document id'
type SdocID string

// DocumentMeta is a document without body
type DocumentMeta struct {
	SdocID    SdocID
	CreatedAt time.Time
	UpdatedAt time.Time
	SourceID  string
}

type Repository interface {
	GetDocumentsMetaBySdocID(ctx context.Context, sdocIDs []SdocID) (map[SdocID]DocumentMeta, error)
}
