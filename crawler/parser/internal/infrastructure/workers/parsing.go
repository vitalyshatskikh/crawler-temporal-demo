package workers

import (
	"context"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/application"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/application/activities"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
)

type ParsingWorker struct {
	client           client.Client
	parserActivities *activities.Parser
}

func NewParsingWorker(
	c client.Client,
	svc *domain.ParsingService,
	advertsRepo application.AdvertsRepository,
) (*ParsingWorker, error) {
	parserActivities, err := activities.NewParser(svc, advertsRepo)
	if err != nil {
		return nil, err
	}
	return &ParsingWorker{client: c, parserActivities: parserActivities}, nil
}

func (p *ParsingWorker) Run(ctx context.Context) error {
	interruptCh := make(chan any)
	go func() {
		<-ctx.Done()
		close(interruptCh)
	}()

	opts := worker.Options{}
	w := worker.New(p.client, "parsing", opts)

	w.RegisterActivityWithOptions(
		p.parserActivities.ParseSearchPage,
		activity.RegisterOptions{
			Name: application.ActivityParseSearchPage,
		},
	)
	w.RegisterActivityWithOptions(
		p.parserActivities.ParseAdvertContent,
		activity.RegisterOptions{
			Name: application.ActivityParseAdvertContent,
		},
	)

	return w.Run(interruptCh)
}
