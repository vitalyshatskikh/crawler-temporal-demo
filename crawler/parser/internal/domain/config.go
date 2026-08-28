package domain

import (
	"fmt"
	"strings"

	"github.com/cbroglie/mustache"
)

type ParsingParam struct {
	Name     string `json:"name"`
	JMESPath string `json:"jmespath"`
	Default  string `json:"default"`
}

type ParsingConfig struct {
	ID                  int
	Name                string
	SourceID            SourceID
	DocumentType        DocumentType
	ExternalURLJMESPath string
	ExternalURLTemplate string
	ContentURLTemplate  string
	Params              []ParsingParam
}

func (c ParsingConfig) Validate() error {
	if c.SourceID == "" {
		return fmt.Errorf("%w: SourceID is required", ErrValidation)
	}
	if c.DocumentType == "" {
		return fmt.Errorf("%w: DocumentType is required", ErrValidation)
	}
	if c.DocumentType == DocumentTypeSearchPage {
		if c.ExternalURLJMESPath == "" {
			return fmt.Errorf("%w: ExternalURLJMESPath is required for search_page", ErrValidation)
		}
	}
	if c.ExternalURLTemplate != "" {
		_, err := mustache.ParseString(c.ExternalURLTemplate)
		if err != nil {
			return fmt.Errorf("%w: ExternalURLTemplate: %w", ErrValidation, err)
		}
	}
	if c.ContentURLTemplate != "" {
		_, err := mustache.ParseString(c.ContentURLTemplate)
		if err != nil {
			return fmt.Errorf("%w: ContentURLTemplate: %w", ErrValidation, err)
		}
	}
	for _, p := range c.Params {
		if strings.HasPrefix(p.Name, "_") {
			return fmt.Errorf("%w: Param.Name '%s' starts with reserved prefix '_'", ErrValidation, p.Name)
		}
	}
	return nil
}
