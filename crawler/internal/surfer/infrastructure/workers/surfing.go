package workers

import (
	"context"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"golang.org/x/sync/errgroup"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/application/advertswf"
)

func RunSurfingWorker(ctx context.Context, c client.Client, surfer *advertswf.Surfer) error {
	eg, egCtx := errgroup.WithContext(ctx)

	interruptCh := make(chan interface{})
	go func() {
		<-egCtx.Done()
		close(interruptCh)
	}()

	eg.Go(func() error {
		opts := worker.Options{}
		w := worker.New(c, advertswf.SurfingTaskQueue, opts)
		w.RegisterWorkflowWithOptions(surfer.SearchAdverts, workflow.RegisterOptions{})
		w.RegisterWorkflowWithOptions(surfer.ProcessSearchBranch, workflow.RegisterOptions{})
		w.RegisterWorkflowWithOptions(surfer.ProcessSearchPage, workflow.RegisterOptions{})
		return w.Run(interruptCh)
	})
	eg.Go(func() error {
		opts := worker.Options{}
		w := worker.New(c, advertswf.AdvertProcessingQueue, opts)
		w.RegisterWorkflowWithOptions(surfer.ProcessAdvert, workflow.RegisterOptions{})
		return w.Run(interruptCh)
	})
	return eg.Wait()
}
