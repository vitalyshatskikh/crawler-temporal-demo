package testutil

import (
	"time"

	"github.com/go-faker/faker/v4"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
)

func SdocIDFactory() domain.SdocID {
	return domain.SdocID(faker.UUIDHyphenated())
}

func SourceIDFactory() domain.SourceID {
	return domain.SourceID(faker.UUIDHyphenated())
}

func DocumentMetaFactory() domain.DocumentMeta {
	return domain.DocumentMeta{
		SdocID:      SdocIDFactory(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		SourceID:    SourceIDFactory(),
		Type:        domain.DocumentTypeSearchPage,
		ExternalURL: faker.URL(),
	}
}

func MustDocumentMeta(url string, sdocID domain.SdocID, sourceID domain.SourceID) domain.DocumentMeta {
	now := time.Now()
	return domain.DocumentMeta{
		SdocID:      sdocID,
		CreatedAt:   now,
		UpdatedAt:   now,
		SourceID:    sourceID,
		Type:        domain.DocumentTypeSearchPage,
		ExternalURL: url,
	}
}

func DocumentFactory() domain.Document {
	return domain.Document{
		DocumentMeta: DocumentMetaFactory(),
		Body:         []byte(`{}`),
	}
}

func ParamFactory() domain.ParsingParam {
	return domain.ParsingParam{
		Name:     faker.Word(),
		JMESPath: faker.Word(),
		Default:  faker.Word(),
	}
}

func ConfigFactory() domain.ParsingConfig {
	return domain.ParsingConfig{
		ID:           1,
		Name:         faker.Word(),
		SourceID:     domain.SourceID(faker.Word()),
		DocumentType: domain.DocumentTypeSearchPage,
		Params:       []domain.ParsingParam{ParamFactory()},
	}
}

func MustSearchPageConfig(sourceID domain.SourceID, params []domain.ParsingParam) domain.ParsingConfig {
	cfg := ConfigFactory()
	cfg.SourceID = sourceID
	cfg.DocumentType = domain.DocumentTypeSearchPage
	if params != nil {
		cfg.Params = params
	}
	return cfg
}

func MustAdvertConfig(sourceID domain.SourceID, params []domain.ParsingParam) domain.ParsingConfig {
	cfg := ConfigFactory()
	cfg.SourceID = sourceID
	cfg.DocumentType = domain.DocumentTypeDownloadedAdvert
	if params != nil {
		cfg.Params = params
	}
	return cfg
}

func ValidSearchPageConfigFactory() domain.ParsingConfig {
	return domain.ParsingConfig{
		ID:           1,
		Name:         "search-config",
		SourceID:     domain.SourceID(faker.Word()),
		DocumentType: domain.DocumentTypeSearchPage,
		Params: []domain.ParsingParam{
			{Name: domain.PropExternalURL, JMESPath: "urls[*]", Default: ""},
		},
	}
}

func ValidAdvertConfigFactory() domain.ParsingConfig {
	return domain.ParsingConfig{
		ID:           2,
		Name:         "advert-config",
		SourceID:     domain.SourceID(faker.Word()),
		DocumentType: domain.DocumentTypeDownloadedAdvert,
		Params: []domain.ParsingParam{
			{Name: domain.PropExternalURL, JMESPath: "url", Default: ""},
		},
	}
}
