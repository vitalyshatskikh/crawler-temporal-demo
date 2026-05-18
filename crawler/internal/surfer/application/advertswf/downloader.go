package advertswf

import (
	"go.temporal.io/sdk/workflow"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/adverts"
)

type Downloader interface {
	DownloadSearchPage(ctx workflow.Context, url string) (adverts.DocumentMeta, error)
	DownloadAdvertContent(ctx workflow.Context, sdocID adverts.SdocID) (adverts.DocumentMeta, error)
}
