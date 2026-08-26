package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type parserKey struct {
	sourceID SourceID
	docType  DocumentType
}

// ParsingService compiles JMESPath parsers from ParsingConfig and applies them
// to Documents. Parsers are cached per (SourceID, DocumentType) pair and are
// never invalidated for the lifetime of the service; configs are treated as
// static. Use a new service instance if config refresh is needed.
type ParsingService struct {
	confRepo ConfigRepository

	mu          sync.RWMutex
	jmesParsers map[parserKey]*JMESParser
	sf          singleflight.Group
}

func NewParsingService(confRepo ConfigRepository) (*ParsingService, error) {
	if confRepo == nil {
		return nil, ErrValidation
	}
	return &ParsingService{
		confRepo:    confRepo,
		jmesParsers: make(map[parserKey]*JMESParser),
	}, nil
}

func (s *ParsingService) ParseSearchPage(ctx context.Context, doc Document) ([]Document, error) {
	if doc.Type != DocumentTypeSearchPage {
		return nil, fmt.Errorf("%w: doc.Type must be DocumentTypeSearchPage", ErrValidation)
	}
	if err := doc.Validate(); err != nil {
		return nil, err
	}

	parser, err := s.getJMESParser(ctx, doc.SourceID, doc.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get search page parser %s/%s: %w", doc.SourceID, doc.SdocID, err)
	}

	parsed, err := parser.Parse(ctx, doc.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse search page %s/%s: %w", doc.SourceID, doc.SdocID, err)
	}

	urls := parsed[PropExternalURL]
	parsedDocs := make([]Document, 0, len(urls))
	now := time.Now()
	for i, u := range urls {
		extURL := fmt.Sprint(u)
		sdocID, err := SdocIDForURL(extURL)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to compute SdocID for %s: %w", ErrValidation, extURL, err)
		}

		docBody := make(map[string]any, len(parsed))
		for key, vals := range parsed {
			if i < len(vals) {
				docBody[key] = vals[i]
			}
		}

		buf := &bytes.Buffer{}
		err = json.NewEncoder(buf).Encode(docBody)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to encode parsed snippet %s/%s: %w",
				doc.SourceID, sdocID, err,
			)
		}

		parsedDoc := Document{
			DocumentMeta: DocumentMeta{
				SdocID:            sdocID,
				CreatedAt:         now,
				UpdatedAt:         now,
				SourceID:          doc.SourceID,
				Type:              DocumentTypeSurfedAdvert,
				ExternalURL:       extURL,
				UpdateIntervalSec: doc.UpdateIntervalSec,
			},
			Body: buf.Bytes(),
		}
		parsedDocs = append(parsedDocs, parsedDoc)
	}

	return parsedDocs, nil
}

func (s *ParsingService) ParseAdvertContent(ctx context.Context, doc Document) (Document, error) {
	if doc.Type != DocumentTypeDownloadedAdvert {
		return Document{}, fmt.Errorf("%w: doc.Type must be DocumentTypeDownloadedAdvert", ErrValidation)
	}
	if err := doc.Validate(); err != nil {
		return Document{}, err
	}

	parser, err := s.getJMESParser(ctx, doc.SourceID, doc.Type)
	if err != nil {
		return Document{}, fmt.Errorf("failed to get advert parser %s/%s: %w", doc.SourceID, doc.SdocID, err)
	}

	parsed, err := parser.Parse(ctx, doc.Body)
	if err != nil {
		return Document{}, fmt.Errorf("failed to parse advert %s/%s: %w", doc.SourceID, doc.SdocID, err)
	}

	buf := &bytes.Buffer{}
	err = json.NewEncoder(buf).Encode(parsed)
	if err != nil {
		return Document{}, fmt.Errorf(
			"failed to encode parsed advert %s/%s: %w",
			doc.SourceID, doc.SdocID, err,
		)
	}

	now := time.Now()
	return Document{
		DocumentMeta: DocumentMeta{
			SdocID:            doc.SdocID,
			CreatedAt:         now,
			UpdatedAt:         now,
			SourceID:          doc.SourceID,
			Type:              DocumentTypeParsedAdvert,
			ExternalURL:       doc.ExternalURL,
			UpdateIntervalSec: doc.UpdateIntervalSec,
		},
		Body: buf.Bytes(),
	}, nil
}

func (s *ParsingService) getJMESParser(
	ctx context.Context,
	sourceID SourceID,
	docType DocumentType,
) (*JMESParser, error) {
	key := parserKey{sourceID: sourceID, docType: docType}

	s.mu.RLock()
	parser, ok := s.jmesParsers[key]
	s.mu.RUnlock()
	if ok {
		return parser, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	sep := "\x00"
	result, sfErr, _ := s.sf.Do(string(sourceID)+sep+string(docType), func() (any, error) {
		cnf, e := s.confRepo.GetConfig(ctx, sourceID, docType)
		if e != nil {
			return nil, e
		}
		if ve := cnf.Validate(); ve != nil {
			return nil, fmt.Errorf("%w: %w", ErrValidation, ve)
		}
		parser, e := NewJMESParser(cnf)
		if e != nil {
			return nil, e
		}
		s.mu.Lock()
		s.jmesParsers[key] = parser
		s.mu.Unlock()
		return parser, nil
	})
	if sfErr != nil {
		return nil, sfErr
	}
	return result.(*JMESParser), nil
}
