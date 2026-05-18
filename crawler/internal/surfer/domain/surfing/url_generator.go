package surfing

import (
	"maps"
	"strconv"
	"sync"

	"github.com/cbroglie/mustache"
)

type URLGenerator struct {
	template *mustache.Template
}

// NewURLGenerator creates new renderer to substitute url parameters using 'mustache' syntax (logicless).
//
// https://mustache.github.io/
func NewURLGenerator(tmpl string) (URLGenerator, error) {
	parsedTmpl, err := mustache.ParseString(tmpl)
	if err != nil {
		return URLGenerator{}, err
	}
	return URLGenerator{
		template: parsedTmpl,
	}, nil
}

func (r URLGenerator) Branch(params TemplateContext) BranchURLGenerator {
	values := make(map[string]string, len(params.Values))
	maps.Copy(values, params.Values)
	return BranchURLGenerator{
		template: r.template,
		params:   values,
		mu:       &sync.Mutex{},
	}
}

type BranchURLGenerator struct {
	template *mustache.Template
	params   map[string]string

	mu *sync.Mutex
}

func (r *BranchURLGenerator) Page(pageNum int) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.params[URLTemplatePageParam] = strconv.Itoa(pageNum)
	return r.template.Render(r.params)
}
