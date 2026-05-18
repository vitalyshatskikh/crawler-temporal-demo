package advertswf_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/application/advertswf"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/adverts"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/parsing"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/internal/surfer/domain/surfing"
)

type TestSuite struct {
	testsuite.WorkflowTestSuite
}

func TestNewSurfer(t *testing.T) {
	t.Run("when_success__then_no_error", func(t *testing.T) {
		// Given
		okConfigRepo := &surfing.DummyConfigRepository{}
		okAdvertsRepo := &adverts.DummyAdvertsRepository{}

		// When
		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)

		// Then
		require.NoError(t, err)
		require.NotNil(t, s)
	})

	t.Run("when_config_repo_is_nil__then_returns_error", func(t *testing.T) {
		// Given
		var configRepo surfing.Repository = nil
		okAdvertsRepo := &adverts.DummyAdvertsRepository{}

		// When
		s, err := advertswf.NewSurfer(configRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)

		// Then
		require.Error(t, err)
		require.Nil(t, s)
		assert.Contains(t, err.Error(), "configRepo is nil")
	})

	t.Run("when_adverts_repo_is_nil__then_returns_error", func(t *testing.T) {
		// Given
		okConfigRepo := &surfing.DummyConfigRepository{}
		var advertsRepo adverts.Repository = nil

		// When
		s, err := advertswf.NewSurfer(okConfigRepo, advertsRepo, advertswf.DefaultSurferConfig)

		// Then
		require.Error(t, err)
		require.Nil(t, s)
		assert.Contains(t, err.Error(), "advertsRepo is nil")
	})
}

func TestSurfer_SearchAdverts(t *testing.T) {
	var surfParams surfing.Params
	err := faker.FakeData(&surfParams)
	require.NoError(t, err)

	surfParams.URLTemplateParams = []surfing.TemplateContext{
		{Values: map[string]string{"cat": faker.Word()}},
		{Values: map[string]string{"cat": faker.Word()}},
		{Values: map[string]string{"cat": faker.Word()}},
	}

	okConfigRepo := &surfing.DummyConfigRepository{
		GetSurfConfigResult: surfParams,
		GetSurfConfigError:  nil,
	}

	okAdvertsRepo := &adverts.DummyAdvertsRepository{}

	t.Run("when_success__then_no_error", func(t *testing.T) {
		// Given
		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		for i := 0; i < len(surfParams.URLTemplateParams); i++ {
			expectedBranchIdx := i
			env.OnWorkflow(s.ProcessSearchBranch, mock.Anything, surfParams, expectedBranchIdx).
				Return(nil).
				Once()
		}

		configName := faker.Word()

		// When
		env.ExecuteWorkflow(s.SearchAdverts, configName)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		require.True(t, env.AssertExpectations(t))
	})

	t.Run("when_config_load_fails__then_returns_error", func(t *testing.T) {
		// Given
		configName := faker.Word()
		configErr := assert.AnError

		failConfigRepo := &surfing.DummyConfigRepository{
			GetSurfConfigResult: surfing.Params{},
			GetSurfConfigError:  configErr,
		}

		s, err := advertswf.NewSurfer(failConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		// When
		env.ExecuteWorkflow(s.SearchAdverts, configName)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.Error(t, env.GetWorkflowError())
		require.Contains(t, env.GetWorkflowError().Error(), "cannot get surfing config")
	})

	t.Run("when_one_branch_fails__then_returns_error", func(t *testing.T) {
		// Given
		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		env.OnWorkflow(s.ProcessSearchBranch, mock.Anything, surfParams, 0).
			Return(nil).
			Once()
		env.OnWorkflow(s.ProcessSearchBranch, mock.Anything, surfParams, 1).
			Return(assert.AnError).
			Once()
		env.OnWorkflow(s.ProcessSearchBranch, mock.Anything, surfParams, 2).
			Return(nil).
			Once()

		configName := faker.Word()

		// When
		env.ExecuteWorkflow(s.SearchAdverts, configName)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.Error(t, env.GetWorkflowError())
		// Temporal does not return original err, so errors.Is/As are unuseful
		errMsg := env.GetWorkflowError().Error()
		require.Contains(t, errMsg, "assert.AnError", "should contain the actual error")
		require.Contains(t, errMsg, "ProcessSearchBranch", "should indicate branch workflow error")
		require.True(t, env.AssertExpectations(t))
	})

	t.Run("when_multiple_branches_fail__then_returns_joined_errors", func(t *testing.T) {
		// Given
		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		brErr1 := errors.New("branch 1 failed")
		brErr2 := errors.New("branch 2 failed")

		env.OnWorkflow(s.ProcessSearchBranch, mock.Anything, surfParams, 0).
			Return(brErr1).
			Once()
		env.OnWorkflow(s.ProcessSearchBranch, mock.Anything, surfParams, 1).
			Return(brErr2).
			Once()
		env.OnWorkflow(s.ProcessSearchBranch, mock.Anything, surfParams, 2).
			Return(nil).
			Once()

		configName := faker.Word()

		// When
		env.ExecuteWorkflow(s.SearchAdverts, configName)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.Error(t, env.GetWorkflowError())
		// Temporal does not return original err, so errors.Is/As are unuseful
		errMsg := env.GetWorkflowError().Error()
		require.Contains(t, errMsg, "branch 1 failed", "should contain first error")
		require.Contains(t, errMsg, "branch 2 failed", "should contain second error")
		require.Contains(t, errMsg, "ProcessSearchBranch", "should indicate branch workflow errors")
		require.True(t, env.AssertExpectations(t))
	})

	t.Run("when_no_url_template_params__then_single_branch_returns_no_error", func(t *testing.T) {
		// Given
		var surfParamsNoBranches surfing.Params
		err := faker.FakeData(&surfParamsNoBranches)
		require.NoError(t, err)

		surfParamsNoBranches.URLTemplateParams = []surfing.TemplateContext{}

		surfConfigRepo := &surfing.DummyConfigRepository{
			GetSurfConfigResult: surfParamsNoBranches,
			GetSurfConfigError:  nil,
		}

		s, err := advertswf.NewSurfer(surfConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		env.OnWorkflow(s.ProcessSearchBranch, mock.Anything, mock.Anything, 0).
			Return(nil).
			Once()

		configName := faker.Word()

		// When
		env.ExecuteWorkflow(s.SearchAdverts, configName)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		require.True(t, env.AssertExpectations(t))
	})

	t.Run("when_single_branch_success__then_no_error", func(t *testing.T) {
		// Given
		var surfParamsSingleBranch surfing.Params
		err := faker.FakeData(&surfParams)
		require.NoError(t, err)

		surfParamsSingleBranch.URLTemplateParams = []surfing.TemplateContext{
			{Values: map[string]string{"cat": faker.Word()}},
		}

		surfConfigRepo := &surfing.DummyConfigRepository{
			GetSurfConfigResult: surfParamsSingleBranch,
			GetSurfConfigError:  nil,
		}

		s, err := advertswf.NewSurfer(surfConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		env.OnWorkflow(s.ProcessSearchBranch, mock.Anything, surfParamsSingleBranch, 0).
			Return(nil).
			Once()

		configName := faker.Word()

		// When
		env.ExecuteWorkflow(s.SearchAdverts, configName)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		require.True(t, env.AssertExpectations(t))
	})
}

func TestSurfer_ProcessSearchBranch(t *testing.T) {
	surfParams := surfing.Params{
		Name:        faker.Word(),
		URLTemplate: "https://example.com/{{cat}}/search?page={{page}}",
		URLTemplateParams: []surfing.TemplateContext{
			{Values: map[string]string{"cat": faker.Word()}},
			{Values: map[string]string{"cat": faker.Word()}},
		},
		MaxPages: 3,
	}

	branchIdx := 0
	branchCategory := surfParams.URLTemplateParams[branchIdx].Values["cat"]

	okAdvertsRepo := &adverts.DummyAdvertsRepository{}
	okConfigRepo := &surfing.DummyConfigRepository{}

	t.Run("when_success__then_all_pages_processed", func(t *testing.T) {
		// Given
		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		for i := range surfParams.MaxPages {
			pageNum := i + 1
			expectedURL := fmt.Sprintf("https://example.com/%s/search?page=%d", branchCategory, pageNum)
			env.OnWorkflow(s.ProcessSearchPage, mock.Anything, surfParams, branchIdx, pageNum, expectedURL).
				Return(nil).
				Once()
		}

		// When
		env.ExecuteWorkflow(s.ProcessSearchBranch, surfParams, branchIdx)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		require.True(t, env.AssertExpectations(t))
	})

	t.Run("when_bad_url_template__then_returns_error", func(t *testing.T) {
		// Given
		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		badSurfParams := surfParams
		badSurfParams.URLTemplate = "x={{#}"

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()
		// When
		env.ExecuteWorkflow(s.ProcessSearchBranch, badSurfParams, branchIdx)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.Error(t, env.GetWorkflowError())
		require.Contains(t, env.GetWorkflowError().Error(), "cannot create url generator")
	})

	t.Run("when_invalid_params__then_returns_validation_error", func(t *testing.T) {
		// Given
		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		invalidSurfParams := surfing.Params{}

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		// When
		env.ExecuteWorkflow(s.ProcessSearchBranch, invalidSurfParams, branchIdx)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.Error(t, env.GetWorkflowError())
		require.Contains(t, env.GetWorkflowError().Error(), "invalid surfing params")
	})

	for _, badBranchIdx := range []int{-1, len(surfParams.URLTemplateParams)} {
		t.Run(
			fmt.Sprintf("when_invalid_branch_idx_%d__then_returns_validation_error", badBranchIdx),
			func(t *testing.T) {
				// Given
				s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
				require.NoError(t, err)

				suite := &TestSuite{}
				env := suite.NewTestWorkflowEnvironment()

				// When
				env.ExecuteWorkflow(s.ProcessSearchBranch, surfParams, badBranchIdx)

				// Then
				require.True(t, env.IsWorkflowCompleted())
				require.Error(t, env.GetWorkflowError())
				require.Contains(t, env.GetWorkflowError().Error(), "invalid branch index")
			},
		)
	}

	for _, testArgs := range []struct {
		name    string
		pageNum int
	}{
		{name: "first", pageNum: 1},
		{name: "middle", pageNum: 2},
		{name: "last", pageNum: 3},
	} {
		t.Run(
			fmt.Sprintf("when_%s_page_fails__then_returns_error_and_stops", testArgs.name),
			func(t *testing.T) {
				// Given
				s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
				require.NoError(t, err)

				suite := &TestSuite{}
				env := suite.NewTestWorkflowEnvironment()

				returnValues := make([]error, testArgs.pageNum)
				returnValues[testArgs.pageNum-1] = assert.AnError
				env.OnWorkflow(s.ProcessSearchPage, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(func(workflow.Context, surfing.Params, int, int, string) error {
						resp := returnValues[0]
						returnValues = returnValues[1:]
						return resp
					}).
					Times(testArgs.pageNum)

				// When
				env.ExecuteWorkflow(s.ProcessSearchBranch, surfParams, branchIdx)

				// Then
				require.True(t, env.IsWorkflowCompleted())
				require.Error(t, env.GetWorkflowError())
				// Verify error message shows correct page number (1-based)
				errMsg := env.GetWorkflowError().Error()
				assert.Contains(t, errMsg, fmt.Sprintf("failed on page %d", testArgs.pageNum), "should show page %d (1-based)", testArgs.pageNum)
				require.True(t, env.AssertExpectations(t))
			},
		)
	}

	t.Run("when_max_pages_is_1__then_processes_single_page", func(t *testing.T) {
		// Given
		singlePageParams := surfParams
		singlePageParams.MaxPages = 1

		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		env.OnWorkflow(s.ProcessSearchPage, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil).
			Once()

		// When
		env.ExecuteWorkflow(s.ProcessSearchBranch, singlePageParams, branchIdx)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		require.True(t, env.AssertExpectations(t))
	})

	t.Run("when_max_pages_is_0__then_processes_no_pages", func(t *testing.T) {
		// Given
		zeroPageParams := surfParams
		zeroPageParams.MaxPages = 0

		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		// When
		env.ExecuteWorkflow(s.ProcessSearchBranch, zeroPageParams, branchIdx)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.Error(t, env.GetWorkflowError())
		require.Contains(t, env.GetWorkflowError().Error(), "invalid surfing params")
	})
}

func TestSurfer_ProcessSearchPage(t *testing.T) {
	surfParams := surfing.Params{
		Name:        faker.Word(),
		URLTemplate: "https://example.com",
		MaxPages:    10,
	}

	branchIdx := 0
	pageNum := 1
	pageURL := faker.URL()

	var pageMeta adverts.DocumentMeta
	err := faker.FakeData(&pageMeta)
	require.NoError(t, err)

	sdocIDs := make([]adverts.SdocID, 3)
	for i := 0; i < len(sdocIDs); i++ {
		sdocIDs[i] = adverts.SdocID(faker.Word())
	}

	documentsMeta := make(map[adverts.SdocID]adverts.DocumentMeta, len(sdocIDs))
	var meta adverts.DocumentMeta
	for _, sdocID := range sdocIDs {
		require.NoError(t, faker.FakeData(&meta))
		meta.SdocID = sdocID
		documentsMeta[sdocID] = meta
	}

	okDownloader := &advertswf.DummyDownloader{
		DownloadSearchPageDocument: pageMeta,
		DownloadSearchPageError:    nil,
	}

	okParser := &parsing.DummyParser{
		ParseSearchPageSdocIDs: sdocIDs,
		ParseSearchPageError:   nil,
	}

	okAdvertsRepo := &adverts.DummyAdvertsRepository{
		GetDocumentsMetaBySdocIDResult: documentsMeta,
		GetDocumentsMetaBySdocIDError:  nil,
	}

	okConfigRepo := &surfing.DummyConfigRepository{
		GetSurfConfigResult: surfParams,
		GetSurfConfigError:  nil,
	}

	t.Run("when_success__then_no_error", func(t *testing.T) {
		// Given
		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		env.RegisterWorkflow(okDownloader.DownloadSearchPage)
		env.RegisterActivity(okParser.ParseSearchPage)
		env.RegisterActivity(okAdvertsRepo.GetDocumentsMetaBySdocID)

		wg := sync.WaitGroup{}
		wg.Add(len(sdocIDs))
		for _, sdocID := range sdocIDs {
			docMetaMatcher := mock.MatchedBy(func(gotMeta adverts.DocumentMeta) bool {
				return gotMeta == adverts.DocumentMeta{
					SdocID:    sdocID,
					CreatedAt: gotMeta.CreatedAt.Truncate(0), // strip monotonic part
					UpdatedAt: gotMeta.UpdatedAt.Truncate(0), // strip monotonic part
					SourceID:  gotMeta.SourceID,
				}
			})
			env.OnWorkflow(s.ProcessAdvert, mock.Anything, surfParams, docMetaMatcher).
				Run(func(args mock.Arguments) {
					wg.Done()
				}).
				Return(nil).
				Once()
		}

		// When
		env.ExecuteWorkflow(s.ProcessSearchPage, surfParams, branchIdx, pageNum, pageURL)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		wg.Wait() // ensure child WF were run
		require.True(t, env.AssertExpectations(t))
	})

	t.Run("when_downloading_fails__then_returns_error", func(t *testing.T) {
		// Given
		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		downloadErr := assert.AnError
		downloader := &advertswf.DummyDownloader{
			DownloadSearchPageError: downloadErr,
		}

		env.RegisterWorkflow(downloader.DownloadSearchPage)

		// When
		env.ExecuteWorkflow(s.ProcessSearchPage, surfParams, branchIdx, pageNum, pageURL)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.Error(t, env.GetWorkflowError())
		assert.Contains(t, env.GetWorkflowError().Error(), "cannot download search page")
	})

	t.Run("when_parsing_fails__then_returns_error", func(t *testing.T) {
		// Given
		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		parseErr := assert.AnError
		parser := &parsing.DummyParser{
			ParseSearchPageError: parseErr,
		}

		env.RegisterWorkflow(okDownloader.DownloadSearchPage)
		env.RegisterActivity(parser.ParseSearchPage)

		// When
		env.ExecuteWorkflow(s.ProcessSearchPage, surfParams, branchIdx, pageNum, pageURL)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.Error(t, env.GetWorkflowError())
		assert.Contains(t, env.GetWorkflowError().Error(), "cannot parse search page")
	})

	t.Run("when_get_documents_meta_fails__then_returns_error", func(t *testing.T) {
		// Given
		getMetaErr := assert.AnError
		advertsRepo := &adverts.DummyAdvertsRepository{
			GetDocumentsMetaBySdocIDError: getMetaErr,
		}

		s, err := advertswf.NewSurfer(okConfigRepo, advertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		env.RegisterWorkflow(okDownloader.DownloadSearchPage)
		env.RegisterActivity(okParser.ParseSearchPage)
		env.RegisterActivity(advertsRepo.GetDocumentsMetaBySdocID)

		// When
		env.ExecuteWorkflow(s.ProcessSearchPage, surfParams, branchIdx, pageNum, pageURL)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.Error(t, env.GetWorkflowError())
		assert.Contains(t, env.GetWorkflowError().Error(), "cannot get documents metadata")
	})

	t.Run("when_success__then_process_advert_fire_and_forget", func(t *testing.T) {
		// Given
		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		env.RegisterWorkflow(okDownloader.DownloadSearchPage)
		env.RegisterWorkflow(okDownloader.DownloadAdvertContent)
		env.RegisterActivity(okParser.ParseSearchPage)
		env.RegisterActivity(okAdvertsRepo.GetDocumentsMetaBySdocID)

		wg := sync.WaitGroup{}
		wg.Add(len(sdocIDs))
		env.OnWorkflow(s.ProcessAdvert, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				wg.Done()
			}).
			Return(assert.AnError).
			Times(len(documentsMeta))

		// When
		env.ExecuteWorkflow(s.ProcessSearchPage, surfParams, branchIdx, pageNum, pageURL)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		// Parent completes successfully despite child workflows having errors
		// This proves fire-and-forget behavior
		wg.Wait() // ensure children were called
		require.True(t, env.AssertExpectations(t))
	})

	t.Run("when_no_adverts_found__then_no_child_workflows", func(t *testing.T) {
		// Given
		advertsRepoWithEmpty := &adverts.DummyAdvertsRepository{
			GetDocumentsMetaBySdocIDResult: map[adverts.SdocID]adverts.DocumentMeta{},
			GetDocumentsMetaBySdocIDError:  nil,
		}

		s, err := advertswf.NewSurfer(okConfigRepo, advertsRepoWithEmpty, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		env.RegisterWorkflow(okDownloader.DownloadSearchPage)
		env.RegisterActivity(okParser.ParseSearchPage)
		env.RegisterActivity(advertsRepoWithEmpty.GetDocumentsMetaBySdocID)

		env.OnWorkflow(s.ProcessAdvert, mock.Anything, mock.Anything).Return(nil).Never()

		// When
		env.ExecuteWorkflow(s.ProcessSearchPage, surfParams, branchIdx, pageNum, pageURL)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		// No child workflows should be started when there are no adverts
		require.True(t, env.AssertExpectations(t))
	})

	t.Run("when_invalid_params__then_returns_validation_error", func(t *testing.T) {
		// Given
		s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
		require.NoError(t, err)

		invalidSurfParams := surfing.Params{}

		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		// When
		env.ExecuteWorkflow(s.ProcessSearchPage, invalidSurfParams, branchIdx, pageNum, pageURL)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.Error(t, env.GetWorkflowError())
		require.Contains(t, env.GetWorkflowError().Error(), "invalid surfing params")
	})
}

func TestSurfer_ProcessAdvert(t *testing.T) {
	okConfigRepo := &surfing.DummyConfigRepository{}
	okAdvertsRepo := &adverts.DummyAdvertsRepository{}
	s, err := advertswf.NewSurfer(okConfigRepo, okAdvertsRepo, advertswf.DefaultSurferConfig)
	require.NoError(t, err)

	surfParams := surfing.Params{
		Name:        faker.Word(),
		URLTemplate: "https://example.com",
		MaxPages:    10,
	}

	var docMeta adverts.DocumentMeta
	err = faker.FakeData(&docMeta)
	require.NoError(t, err)

	okDownloader := &advertswf.DummyDownloader{
		DownloadAdvertContentDocument: docMeta,
		DownloadAdvertContentError:    nil,
	}

	okParser := &parsing.DummyParser{
		ParseAdvertContentError: nil,
	}

	t.Run("when_success__then_no_error", func(t *testing.T) {
		// Given
		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		env.RegisterWorkflow(okDownloader.DownloadAdvertContent)
		env.RegisterActivity(okParser.ParseAdvertContent)

		// When
		env.ExecuteWorkflow(s.ProcessAdvert, surfParams, docMeta)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
	})

	t.Run("when_parsing_fails__then_returns_error", func(t *testing.T) {
		// Given
		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		parseErr := assert.AnError
		parser := &parsing.DummyParser{
			ParseAdvertContentError: parseErr,
		}

		env.RegisterWorkflow(okDownloader.DownloadAdvertContent)
		env.RegisterActivity(parser.ParseAdvertContent)

		// When
		env.ExecuteWorkflow(s.ProcessAdvert, surfParams, docMeta)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.Error(t, env.GetWorkflowError())
		assert.Contains(t, env.GetWorkflowError().Error(), "cannot parse advert")
	})

	t.Run("when_invalid_params__then_returns_validation_error", func(t *testing.T) {
		// Given
		suite := &TestSuite{}
		env := suite.NewTestWorkflowEnvironment()

		invalidSurfParams := surfing.Params{}

		// When
		env.ExecuteWorkflow(s.ProcessAdvert, invalidSurfParams, docMeta)

		// Then
		require.True(t, env.IsWorkflowCompleted())
		require.Error(t, env.GetWorkflowError())
		require.Contains(t, env.GetWorkflowError().Error(), "invalid surfing params")
	})
}
