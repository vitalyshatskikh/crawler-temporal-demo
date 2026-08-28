package domain

import "errors"

var (
	ErrParsingFailed = errors.New("parsing failed")
	ErrValidation    = errors.New("validation failed")
	ErrNotFound      = errors.New("item not found")
)
