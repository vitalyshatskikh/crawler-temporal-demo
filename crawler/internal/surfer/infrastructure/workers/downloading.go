package workers

import (
	"context"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"golang.org/x/sync/errgroup"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/application/advertswf"
)

func RunDownloadingWorker(ctx context.Context, c client.Client, downloader advertswf.Downloader) error {
	eg, egCtx := errgroup.WithContext(ctx)

	interruptCh := make(chan interface{})
	go func() {
		<-egCtx.Done()
		close(interruptCh)
	}()

	eg.Go(func() error {
		opts := worker.Options{}
		w := worker.New(c, advertswf.SearchPageDownloadingQueue, opts)
		w.RegisterWorkflowWithOptions(downloader.DownloadSearchPage, workflow.RegisterOptions{})
		return w.Run(interruptCh)
	})
	eg.Go(func() error {
		opts := worker.Options{}
		w := worker.New(c, advertswf.AdvertDownloadingQueue, opts)
		w.RegisterWorkflowWithOptions(downloader.DownloadAdvertContent, workflow.RegisterOptions{})
		return w.Run(interruptCh)
	})
	return eg.Wait()
}
