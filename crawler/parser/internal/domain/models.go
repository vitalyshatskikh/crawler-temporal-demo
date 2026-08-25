// Package domain defines main app entities and core services
//
// Note: keep models in sync with 'crawler/surfer/domain/adverts/models.py'
package domain

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

type SdocID string

type SourceID string

type DocumentType string

const (
	DocumentTypeSearchPage       DocumentType = "search_page"
	DocumentTypeSurfedAdvert     DocumentType = "surfed_advert"
	DocumentTypeDownloadedAdvert DocumentType = "downloaded_advert"
	DocumentTypeParsedAdvert     DocumentType = "parsed_advert"
)

type DocumentMeta struct {
	SdocID      SdocID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	SourceID    SourceID
	Type        DocumentType
	ExternalURL string
}

func (d DocumentMeta) Validate() error {
	if d.SourceID == "" {
		return fmt.Errorf("%w: SourceID is required", ErrValidation)
	}
	if d.SdocID == "" {
		return fmt.Errorf("%w: SdocID is required", ErrValidation)
	}
	if d.ExternalURL == "" {
		return fmt.Errorf("%w: ExternalURL is required", ErrValidation)
	}
	if d.CreatedAt.IsZero() {
		return fmt.Errorf("%w: CreatedAt is required", ErrValidation)
	}
	if d.UpdatedAt.IsZero() {
		return fmt.Errorf("%w: UpdatedAt is required", ErrValidation)
	}
	if d.UpdatedAt.Before(d.CreatedAt) {
		return fmt.Errorf("%w: UpdatedAt must not be before CreatedAt", ErrValidation)
	}
	return nil
}

func SdocIDForURL(urlStr string) (SdocID, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrValidation, err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%w: no host in URL", ErrValidation)
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.Fragment = ""
	keys := make([]string, 0, len(u.Query()))
	for k := range u.Query() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rebuilt := url.Values{}
	for _, k := range keys {
		rebuilt[k] = u.Query()[k]
	}
	u.RawQuery = rebuilt.Encode()
	if u.Path != "/" {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}
	normalized := u.String()
	hash := md5.Sum([]byte(normalized))
	return SdocID(hex.EncodeToString(hash[:])), nil
}

type Document struct {
	DocumentMeta
	Body []byte
}
