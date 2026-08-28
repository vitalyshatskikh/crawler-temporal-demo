package domain

import (
	"errors"

	"github.com/cbroglie/mustache"
)

var ErrMustache = errors.New("mustache error")

func RenderTemplate(tpl string, ctx map[string]any) string {
	result, err := mustache.Render(tpl, ctx)
	if err != nil {
		return ""
	}
	return result
}
