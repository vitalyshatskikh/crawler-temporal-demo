package advertswf

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	wf "go.temporal.io/sdk/workflow"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/adverts"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/parsing"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/surfing"
)

type Surfer struct {
	configRepo  surfing.Repository
	advertsRepo adverts.Repository

	validator *validator.Validate

	processBranchOptions         wf.ChildWorkflowOptions
	processSearchPageOptions     wf.ChildWorkflowOptions
	downloadSearchPageOptions    wf.ChildWorkflowOptions
	processAdvertOptions         wf.ChildWorkflowOptions
	downloadAdvertContentOptions wf.ChildWorkflowOptions

	getSurfConfigOptions    wf.LocalActivityOptions
	getDocumentsMetaOptions wf.LocalActivityOptions

	parseSearchPageOptions    wf.ActivityOptions
	parseAdvertContentOptions wf.ActivityOptions
}

func NewSurfer(configRepo surfing.Repository, advertsRepo adverts.Repository, config SurferConfig) (*Surfer, error) {
	if configRepo == nil {
		return nil, fmt.Errorf("configRepo is nil")
	}
	if advertsRepo == nil {
		return nil, fmt.Errorf("advertsRepo is nil")
	}
	v := validator.New()
	if err := v.Struct(config); err != nil {
		return nil, err
	}
	surfer := &Surfer{
		configRepo:  configRepo,
		advertsRepo: advertsRepo,
		validator:   v,
		processBranchOptions: wf.ChildWorkflowOptions{
			WorkflowExecutionTimeout: config.ProcessBranchTimeout,
		},
		processSearchPageOptions: wf.ChildWorkflowOptions{
			WorkflowExecutionTimeout: config.ProcessSearchPageTimeout,
		},
		downloadSearchPageOptions: wf.ChildWorkflowOptions{
			TaskQueue:                SearchPageDownloadingQueue,
			WorkflowExecutionTimeout: config.DownloadSearchPageTimeout,
		},
		processAdvertOptions: wf.ChildWorkflowOptions{
			TaskQueue:                AdvertProcessingQueue,
			WorkflowExecutionTimeout: config.ProcessAdvertTimeout,
			ParentClosePolicy:        enums.PARENT_CLOSE_POLICY_ABANDON, // fire-and-forget},
		},
		downloadAdvertContentOptions: wf.ChildWorkflowOptions{
			TaskQueue:                AdvertDownloadingQueue,
			WorkflowExecutionTimeout: config.DownloadAdvertContentTimeout,
		},
		getSurfConfigOptions: wf.LocalActivityOptions{
			ScheduleToCloseTimeout: config.RepoRequestTimeout,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts:    config.RepoRequestRetry.MaxAttempts,
				InitialInterval:    config.RepoRequestRetry.InitInterval,
				MaximumInterval:    config.RepoRequestRetry.MaxInterval,
				BackoffCoefficient: config.RepoRequestRetry.BackoffCoefficient,
			},
		},
		getDocumentsMetaOptions: wf.LocalActivityOptions{
			ScheduleToCloseTimeout: config.RepoRequestTimeout,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts:    config.RepoRequestRetry.MaxAttempts,
				InitialInterval:    config.RepoRequestRetry.InitInterval,
				MaximumInterval:    config.RepoRequestRetry.MaxInterval,
				BackoffCoefficient: config.RepoRequestRetry.BackoffCoefficient,
			},
		},
		parseSearchPageOptions: wf.ActivityOptions{
			TaskQueue:           SearchPageParsingQueue,
			StartToCloseTimeout: config.ProcessSearchPageTimeout,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts:    config.ParseSearchPageRetry.MaxAttempts,
				InitialInterval:    config.ParseSearchPageRetry.InitInterval,
				MaximumInterval:    config.ParseSearchPageRetry.MaxInterval,
				BackoffCoefficient: config.ParseSearchPageRetry.BackoffCoefficient,
				NonRetryableErrorTypes: []string{
					"ErrNotFound",
					"ErrValidationFailed",
				},
			},
		},
		parseAdvertContentOptions: wf.ActivityOptions{
			TaskQueue:           AdvertParsingQueue,
			StartToCloseTimeout: config.ParseAdvertContentTimeout,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts:    config.ParseAdvertContentRetry.MaxAttempts,
				InitialInterval:    config.ParseAdvertContentRetry.InitInterval,
				MaximumInterval:    config.ParseAdvertContentRetry.MaxInterval,
				BackoffCoefficient: config.ParseAdvertContentRetry.BackoffCoefficient,
				NonRetryableErrorTypes: []string{
					"ErrNotFound",
					"ErrValidationFailed",
				},
			},
		},
	}
	return surfer, nil
}

// SearchAdverts is an entrypoint that initiates crawling process:
//   - reads config from repository
//   - creates 'branches' from config: target search requests with different parameters
//   - handles branches concurrently
//   - returns OK if ALL branches OK, err otherwise
func (s *Surfer) SearchAdverts(ctx wf.Context, configName string) error {
	logger := wf.GetLogger(ctx)
	logger.Info("starting SearchAdverts", "configName", configName)
	defer logger.Info("completed SearchAdverts", "configName", configName)

	surfParams, err := s.getSurfConfig(ctx, configName)
	if err != nil {
		return fmt.Errorf("cannot get surfing config id=%s: %w", configName, err)
	}

	if len(surfParams.URLTemplateParams) == 0 {
		// single branch with empty context
		surfParams.URLTemplateParams = []surfing.TemplateContext{
			{Values: make(map[string]string)},
		}
	}

	branches := make([]wf.Future, len(surfParams.URLTemplateParams))
	for i := 0; i < len(surfParams.URLTemplateParams); i++ {
		wfID := fmt.Sprintf("ProcessSearchBranch_%s_branch%d", surfParams.Name, i)
		wfOpts := s.processBranchOptions
		wfOpts.WorkflowID = wfID
		branches[i] = wf.ExecuteChildWorkflow(
			wf.WithChildOptions(ctx, wfOpts),
			s.ProcessSearchBranch,
			surfParams,
			i,
		)
	}

	branchErrors := make([]error, 0, len(surfParams.URLTemplateParams))
	for _, branch := range branches {
		branchErrors = append(branchErrors, branch.Get(ctx, nil))
	}

	return errors.Join(branchErrors...)
}

// ProcessSearchBranch 'scrolls' the pages in given branch: creates page URL and executes page processing WF.
// Stops branch processing if page fails.
func (s *Surfer) ProcessSearchBranch(ctx wf.Context, surfParams surfing.Params, branchIdx int) error {
	logger := wf.GetLogger(ctx)
	logger.Info("starting ProcessSearchBranch", "configName", surfParams.Name, "branchIdx", branchIdx)
	defer logger.Info("completed ProcessSearchBranch", "configName", surfParams.Name, "branchIdx", branchIdx)

	if err := s.validator.Struct(&surfParams); err != nil {
		return fmt.Errorf("invalid surfing params: %w", err)
	}
	if branchIdx < 0 || branchIdx >= len(surfParams.URLTemplateParams) {
		return fmt.Errorf("invalid branch index: %d", branchIdx)
	}

	urlGen, err := surfing.NewURLGenerator(surfParams.URLTemplate)
	if err != nil {
		return fmt.Errorf("cannot create url generator: %w", err)
	}
	branchGen := urlGen.Branch(surfParams.URLTemplateParams[branchIdx])

	for i := range surfParams.MaxPages {
		pageNum := i + 1
		actualUrl, err := branchGen.Page(pageNum)
		if err != nil {
			return fmt.Errorf("cannot generate page URL: %w", err)
		}

		wfID := fmt.Sprintf("ProcessSearchPage_%s_branch%d_page%d", surfParams.Name, branchIdx, pageNum)
		wfOpts := s.processSearchPageOptions
		wfOpts.WorkflowID = wfID
		err = wf.ExecuteChildWorkflow(
			wf.WithChildOptions(ctx, wfOpts),
			s.ProcessSearchPage,
			surfParams,
			branchIdx,
			pageNum,
			actualUrl,
		).Get(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed on page %d: %w", pageNum, err)
		}
	}

	return nil
}

// ProcessSearchPage handles a search page from given branch:
//   - download page
//   - parse advert snippets and returns sdocIDs
//   - get advert meta from repository
//   - starts processing for each advert (fire and forget)
func (s *Surfer) ProcessSearchPage(
	ctx wf.Context,
	surfParams surfing.Params,
	branchIdx int,
	pageNum int,
	pageURL string,
) error {
	logger := wf.GetLogger(ctx)
	logger.Info("starting ProcessSearchPage", "configName", surfParams.Name, "pageURL", pageURL)
	defer logger.Info("completed ProcessSearchPage", "configName", surfParams.Name, "pageURL", pageURL)

	if err := s.validator.Struct(&surfParams); err != nil {
		return fmt.Errorf("invalid surfing params: %w", err)
	}

	wfID := fmt.Sprintf("DownloadSearchPage_%s_branch%d_page%d", surfParams.Name, branchIdx, pageNum)
	pageMeta, err := s.downloadSearchPage(ctx, pageURL, wfID)
	if err != nil {
		return fmt.Errorf("cannot download search page: %w", err)
	}

	sdocIDs, err := s.parseSearchPage(ctx, pageMeta)
	if err != nil {
		return fmt.Errorf("cannot parse search page: %w", err)
	}

	documentsMeta, err := s.getDocumentsMeta(ctx, sdocIDs)
	if err != nil {
		return fmt.Errorf("cannot get documents metadata: %w", err)
	}

	for sdocID, docMeta := range documentsMeta {
		wfID = fmt.Sprintf("ProcessAdvert_%s_sdocid%s", surfParams.Name, sdocID)
		wfOpts := s.processAdvertOptions
		wfOpts.WorkflowID = wfID
		wf.ExecuteChildWorkflow(
			wf.WithChildOptions(ctx, wfOpts),
			s.ProcessAdvert,
			surfParams,
			docMeta,
		)
	}

	return nil
}

// ProcessAdvert handles a single advert by its sdocID: download -> parse -> index
func (s *Surfer) ProcessAdvert(ctx wf.Context, surfParams surfing.Params, docMeta adverts.DocumentMeta) error {
	logger := wf.GetLogger(ctx)
	logger.Info("starting ProcessSearch", "configName", surfParams.Name, "sdocID", docMeta.SdocID)
	defer logger.Info("completed ProcessSearch", "configName", surfParams.Name, "sdocID", docMeta.SdocID)

	if err := s.validator.Struct(&surfParams); err != nil {
		return fmt.Errorf("invalid surfing params: %w", err)
	}

	wfID := fmt.Sprintf("DownloadAdvertContent_%s_sdocid%s", surfParams.Name, docMeta.SdocID)
	_, err := s.downloadAdvertContent(ctx, docMeta, wfID)
	if err != nil {
		return fmt.Errorf("cannot download advert: %w", err)
	}

	err = s.parseAdvertContent(ctx, docMeta)
	if err != nil {
		return fmt.Errorf("cannot parse advert: %w", err)
	}

	// TODO add indexing stage

	return nil
}

func (s *Surfer) getSurfConfig(ctx wf.Context, configName string) (surfing.Params, error) {
	var surfParams surfing.Params
	err := wf.ExecuteLocalActivity(
		wf.WithLocalActivityOptions(ctx, s.getSurfConfigOptions),
		s.configRepo.GetSurfConfig,
		configName,
	).Get(ctx, &surfParams)
	return surfParams, err
}

func (s *Surfer) getDocumentsMeta(
	ctx wf.Context,
	sdocIDs []adverts.SdocID,
) (map[adverts.SdocID]adverts.DocumentMeta, error) {
	var documentsMeta map[adverts.SdocID]adverts.DocumentMeta
	err := wf.ExecuteLocalActivity(
		wf.WithLocalActivityOptions(ctx, s.getDocumentsMetaOptions),
		s.advertsRepo.GetDocumentsMetaBySdocID,
		sdocIDs,
	).Get(ctx, &documentsMeta)
	return documentsMeta, err
}

func (s *Surfer) downloadSearchPage(ctx wf.Context, pageURL, wfID string) (adverts.DocumentMeta, error) {
	wfOps := s.downloadSearchPageOptions
	wfOps.WorkflowID = wfID
	var pageMeta adverts.DocumentMeta
	err := wf.ExecuteChildWorkflow(
		wf.WithChildOptions(ctx, wfOps),
		Downloader.DownloadSearchPage,
		pageURL,
	).Get(ctx, &pageMeta)
	return pageMeta, err
}

func (s *Surfer) downloadAdvertContent(
	ctx wf.Context,
	docMeta adverts.DocumentMeta,
	wfID string,
) (adverts.DocumentMeta, error) {
	wfOpts := s.downloadAdvertContentOptions
	wfOpts.WorkflowID = wfID
	var downloadedDocMeta adverts.DocumentMeta
	err := wf.ExecuteChildWorkflow(
		wf.WithChildOptions(ctx, wfOpts),
		Downloader.DownloadAdvertContent,
		docMeta.SdocID,
	).Get(ctx, &downloadedDocMeta)
	return downloadedDocMeta, err
}

func (s *Surfer) parseSearchPage(ctx wf.Context, pageMeta adverts.DocumentMeta) ([]adverts.SdocID, error) {
	var sdocIDs []adverts.SdocID
	err := wf.ExecuteActivity(
		wf.WithActivityOptions(ctx, s.parseSearchPageOptions),
		parsing.Parser.ParseSearchPage,
		pageMeta,
	).Get(ctx, &sdocIDs)
	return sdocIDs, err
}

func (s *Surfer) parseAdvertContent(ctx wf.Context, docMeta adverts.DocumentMeta) error {
	err := wf.ExecuteActivity(
		wf.WithActivityOptions(ctx, s.parseAdvertContentOptions),
		parsing.Parser.ParseAdvertContent,
		docMeta,
	).Get(ctx, nil)
	return err
}
