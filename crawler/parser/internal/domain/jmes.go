package domain

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmespath/go-jmespath"
)

type jmesProp struct {
	expr       *jmespath.JMESPath
	defaultVal string
}

type JMESParser struct {
	cnf   ParsingConfig
	props map[string]jmesProp
}

func NewJMESParser(cnf ParsingConfig) (*JMESParser, error) {
	if len(cnf.Params) == 0 {
		return nil, fmt.Errorf("%w: empty Params", ErrValidation)
	}
	props := make(map[string]jmesProp, len(cnf.Params))
	for _, param := range cnf.Params {
		prop, err := jmespath.Compile(param.JMESPath)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: failed to compile JMES path for %s/%s/%s: %w",
				ErrValidation, cnf.SourceID, cnf.DocumentType, param.Name, err,
			)
		}
		props[param.Name] = jmesProp{expr: prop, defaultVal: param.Default}
	}
	return &JMESParser{cnf: cnf, props: props}, nil
}

func (p *JMESParser) Parse(ctx context.Context, body []byte) (map[string][]any, error) {
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParsingFailed, err)
	}

	result := make(map[string][]any, len(p.cnf.Params))
	for name, prop := range p.props {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			parsed, err := prop.expr.Search(data)
			if err != nil {
				return nil, fmt.Errorf(
					"%w: failed to parse prop %s/%s/%s: %w",
					ErrParsingFailed, p.cnf.SourceID, p.cnf.DocumentType, name, err,
				)
			}
			switch val := parsed.(type) {
			case []any:
				result[name] = val
			case nil:
				result[name] = []any{prop.defaultVal}
			case any:
				result[name] = []any{val}
			}
		}
	}
	return result, nil
}
