package activities

import (
	"context"
	"fmt"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/application"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
)

type Parser struct {
	svc  *domain.ParsingService
	repo application.AdvertsRepository
}

func NewParser(svc *domain.ParsingService, repo application.AdvertsRepository) (*Parser, error) {
	if svc == nil || repo == nil {
		return nil, domain.ErrValidation
	}
	return &Parser{svc: svc, repo: repo}, nil
}

func (p *Parser) ParseSearchPage(ctx context.Context, meta domain.DocumentMeta) ([]domain.DocumentMeta, error) {
	if meta.Type != domain.DocumentTypeSearchPage {
		return nil, fmt.Errorf("%w: meta.Type must be DocumentTypeSearchPage", domain.ErrValidation)
	}
	doc, err := p.repo.GetDocument(ctx, meta.SdocID, meta.SourceID, meta.Type)
	if err != nil {
		return nil, fmt.Errorf("get document %s/%s: %w", meta.SourceID, meta.SdocID, err)
	}
	parsed, err := p.svc.ParseSearchPage(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("parse search page %s/%s: %w", meta.SourceID, meta.SdocID, err)
	}
	metas := make([]domain.DocumentMeta, 0, len(parsed))
	for _, d := range parsed {
		if err := p.repo.SaveDocument(ctx, d); err != nil {
			return nil, fmt.Errorf("save document %s/%s: %w", meta.SourceID, d.SdocID, err)
		}
		metas = append(metas, d.DocumentMeta)
	}
	return metas, nil
}

func (p *Parser) ParseAdvertContent(ctx context.Context, meta domain.DocumentMeta) error {
	if meta.Type != domain.DocumentTypeDownloadedAdvert {
		return fmt.Errorf("%w: meta.Type must be DocumentTypeDownloadedAdvert", domain.ErrValidation)
	}
	doc, err := p.repo.GetDocument(ctx, meta.SdocID, meta.SourceID, meta.Type)
	if err != nil {
		return fmt.Errorf("get document %s/%s: %w", meta.SourceID, meta.SdocID, err)
	}
	parsed, err := p.svc.ParseAdvertContent(ctx, doc)
	if err != nil {
		return fmt.Errorf("parse advert %s/%s: %w", meta.SourceID, meta.SdocID, err)
	}
	if err := p.repo.SaveDocument(ctx, parsed); err != nil {
		return fmt.Errorf("save document %s/%s: %w", meta.SourceID, meta.SdocID, err)
	}
	return nil
}
