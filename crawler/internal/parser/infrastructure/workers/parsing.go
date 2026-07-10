package workers

import (
	"context"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"golang.org/x/sync/errgroup"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/application/advertswf"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/parsing"
)

func RunParsingWorker(ctx context.Context, c client.Client, parser parsing.Parser) error {
	eg, egCtx := errgroup.WithContext(ctx)

	interruptCh := make(chan interface{})
	go func() {
		<-egCtx.Done()
		close(interruptCh)
	}()

	cpuOpts := worker.Options{}

	eg.Go(func() error {
		w := worker.New(c, advertswf.SearchPageParsingQueue, cpuOpts)
		w.RegisterActivityWithOptions(parser.ParseSearchPage, activity.RegisterOptions{})
		return w.Run(interruptCh)
	})
	eg.Go(func() error {
		w := worker.New(c, advertswf.AdvertParsingQueue, cpuOpts)
		w.RegisterActivityWithOptions(parser.ParseAdvertContent, activity.RegisterOptions{})
		return w.Run(interruptCh)
	})
	return eg.Wait()
}
