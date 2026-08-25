package domain

import "fmt"

const (
	PropExternalURL = "external_url"
)

type ParsingParam struct {
	Name     string `json:"name"`
	JMESPath string `json:"jmespath"`
	Default  string `json:"default"`
}

type ParsingConfig struct {
	ID           int
	Name         string
	SourceID     SourceID
	DocumentType DocumentType
	Params       []ParsingParam
}

func (c ParsingConfig) Validate() error {
	if c.SourceID == "" {
		return fmt.Errorf("%w: SourceID is required", ErrValidation)
	}
	if c.DocumentType == "" {
		return fmt.Errorf("%w: DocumentType is required", ErrValidation)
	}
	if len(c.Params) < 1 {
		return fmt.Errorf("%w: at least one ParsingParam required", ErrValidation)
	}
	if c.DocumentType == DocumentTypeSearchPage {
		hasExternalURL := false
		for _, p := range c.Params {
			if p.Name == PropExternalURL {
				hasExternalURL = true
				break
			}
		}
		if !hasExternalURL {
			return fmt.Errorf("%w: Params must have '%s' expression", ErrValidation, PropExternalURL)
		}
	}
	return nil
}
