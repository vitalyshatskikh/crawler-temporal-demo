package testutil

import (
	"time"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
)

func MustSearchPageDocument(sourceID domain.SourceID, url string) domain.Document {
	sdocID, err := domain.SdocIDForURL(url)
	if err != nil {
		panic("MustSearchPageDocument: " + err.Error())
	}
	now := time.Now()
	return domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            sdocID,
			CreatedAt:         now,
			UpdatedAt:         now,
			SourceID:          sourceID,
			Type:              domain.DocumentTypeSearchPage,
			ExternalURL:       url,
			ContentURL:        "",
			UpdateIntervalSec: 86400,
		},
		Body: []byte("{}"),
	}
}

func MustDownloadedAdvertDocument(sourceID domain.SourceID, sdocID domain.SdocID, url string) domain.Document {
	now := time.Now()
	return domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            sdocID,
			CreatedAt:         now,
			UpdatedAt:         now,
			SourceID:          sourceID,
			Type:              domain.DocumentTypeDownloadedAdvert,
			ExternalURL:       url,
			ContentURL:        "",
			UpdateIntervalSec: 86400,
		},
		Body: []byte("{}"),
	}
}

func MustSurfedAdvertMeta(sourceID domain.SourceID, sdocID domain.SdocID, url string) domain.DocumentMeta {
	now := time.Now()
	return domain.DocumentMeta{
		SdocID:            sdocID,
		CreatedAt:         now,
		UpdatedAt:         now,
		SourceID:          sourceID,
		Type:              domain.DocumentTypeSurfedAdvert,
		ExternalURL:       url,
		ContentURL:        "",
		UpdateIntervalSec: 86400,
	}
}
