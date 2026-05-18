package advertswf

import (
	"go.temporal.io/sdk/workflow"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/adverts"
)

var (
	_ Downloader = (*DummyDownloader)(nil)
)

type DummyDownloader struct {
	DownloadSearchPageDocument    adverts.DocumentMeta
	DownloadSearchPageError       error
	DownloadAdvertContentDocument adverts.DocumentMeta
	DownloadAdvertContentError    error
}

func (d *DummyDownloader) DownloadSearchPage(_ workflow.Context, _ string) (adverts.DocumentMeta, error) {
	return d.DownloadSearchPageDocument, d.DownloadSearchPageError
}

func (d *DummyDownloader) DownloadAdvertContent(_ workflow.Context, _ adverts.SdocID) (adverts.DocumentMeta, error) {
	return d.DownloadAdvertContentDocument, d.DownloadAdvertContentError
}
