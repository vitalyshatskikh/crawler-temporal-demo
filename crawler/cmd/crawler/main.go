package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	parsew "github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/parser/infrastructure/workers"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/application/advertswf"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/adverts"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/parsing"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/infrastructure/repositories"
	surfw "github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/infrastructure/workers"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
)

var (
	mode = flag.String("m", "", "surfer|parser to run specific worker")
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	flag.Parse()

	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		return err
	}
	defer c.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	switch *mode {
	case "surfer":
		return runSurfer(ctx, c)
	case "parser":
		return runParser(ctx, c)
	default:
		// run all
		eg, egCtx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			return runSurfer(egCtx, c)
		})
		eg.Go(func() error {
			return runParser(egCtx, c)
		})
		return eg.Wait()
	}
}

func runSurfer(ctx context.Context, c client.Client) error {
	configRepo := &repositories.PGConfigRepository{}
	downloader := &advertswf.DummyDownloader{}
	advertRepo := &adverts.DummyAdvertsRepository{}

	surfer, err := advertswf.NewSurfer(configRepo, advertRepo, advertswf.DefaultSurferConfig)
	if err != nil {
		return err
	}

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return surfw.RunSurfingWorker(egCtx, c, surfer)
	})
	eg.Go(func() error {
		return surfw.RunDownloadingWorker(egCtx, c, downloader)
	})

	err = eg.Wait()

	// shutdown gracefully

	return err
}

func runParser(ctx context.Context, c client.Client) error {
	parser := &parsing.DummyParser{}
	return parsew.RunParsingWorker(ctx, c, parser)
}
