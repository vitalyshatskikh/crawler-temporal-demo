package parsing

import (
	"context"
)

type AdvertProp struct {
	Name     string
	JMESPath string
}

type ParseSearchPageParams struct {
	ID               int
	DocumentSourceID string
	SnippetJMESPath  string
	SnipperProps     []AdvertProp
}

type ParseAdvertParams struct {
	ID               int
	DocumentSourceID string
	Props            []AdvertProp
}

type Repository interface {
	GetParseSearchPageConfig(ctx context.Context, docSourceId string) (ParseSearchPageParams, error)
	GetParseAdvertConfig(ctx context.Context, docSourceId string) (ParseAdvertParams, error)
}
