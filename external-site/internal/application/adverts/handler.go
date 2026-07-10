package adverts

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/vitalyshatskikh/crawler-temporal-demo/external-site/internal/application/adverts/gen"
	"github.com/vitalyshatskikh/crawler-temporal-demo/external-site/internal/domain"
)

var _ gen.Handler = (*Handler)(nil)

const (
	DefaultSearchPageSize = 10
	DefaultSearchPageNum  = 1
)

type Handler struct {
	logger  *zap.Logger
	service *domain.AdvertsCRUDService
}

func NewHandler(logger *zap.Logger, service *domain.AdvertsCRUDService) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h Handler) GetAdvert(ctx context.Context, params gen.GetAdvertParams) (gen.GetAdvertRes, error) {
	advert, err := h.service.GetAdvert(ctx, params.Region, params.ID)
	if err != nil {
		return nil, err
	}
	return &gen.AdvertDetail{
		Title:       advert.Title,
		Description: advert.Description,
		Price:       advert.Price,
		PubDate:     advert.PubDate,
	}, nil
}

func (h Handler) SearchAdverts(ctx context.Context, params gen.SearchAdvertsParams) (*gen.SearchResponse, error) {
	adverts, err := h.service.SearchAdverts(
		ctx,
		domain.AdvertSearchParams{
			Region:   params.Region,
			PageSize: params.Size.Or(DefaultSearchPageSize),
			PageNum:  params.Page.Or(DefaultSearchPageNum),
		},
	)
	if err != nil {
		return nil, err
	}
	items := make([]gen.AdvertListItem, 0, len(adverts))
	for _, adv := range adverts {
		advURL := fmt.Sprintf("/adverts/%s/adverts/%s", adv.Region, adv.ID)
		items = append(items, gen.AdvertListItem{Title: adv.Title, URL: advURL})
	}
	return &gen.SearchResponse{
		Page:     0,
		Total:    0,
		NextPage: gen.OptNilString{},
		Adverts:  items,
	}, nil
}

func (h Handler) UpsertAdvert(
	ctx context.Context,
	req *gen.AdvertDetail,
	params gen.UpsertAdvertParams,
) (gen.UpsertAdvertRes, error) {
	created, err := h.service.UpsertAdvert(
		ctx,
		domain.Advert{
			AdvertIdentity: domain.AdvertIdentity{ID: params.ID, Region: params.Region},
			Title:          req.Title,
			Description:    req.Description,
			Price:          req.Price,
			PubDate:        req.PubDate,
		},
	)
	if err != nil {
		return nil, err
	}
	if created {
		return (*gen.UpsertAdvertCreated)(req), nil
	}
	return (*gen.UpsertAdvertOK)(req), nil
}

func (h Handler) DeleteAdvert(ctx context.Context, params gen.DeleteAdvertParams) error {
	err := h.service.DeleteAdvert(ctx, params.Region, params.ID)
	if err != nil {
		return err
	}
	return nil
}

func (h Handler) NewError(_ context.Context, err error) *gen.ErrorStatusCode {
	if err == nil {
		return nil
	}

	detail := map[string]string{"cause": err.Error()}

	switch {
	case errors.Is(err, domain.ErrValidation):
		return newErrorResponse(http.StatusBadRequest, "validation error", detail)
	case errors.Is(err, domain.ErrNotFound):
		return newErrorResponse(http.StatusNotFound, "internal server error", detail)
	default:
		return newErrorResponse(http.StatusInternalServerError, "internal server error", detail)
	}
}

func newErrorResponse(status int, msg string, detail map[string]string) *gen.ErrorStatusCode {
	return &gen.ErrorStatusCode{
		StatusCode: status,
		Response: gen.Error{
			Error: msg,
			Detail: gen.OptErrorDetail{
				Set:   true,
				Value: detail,
			},
		},
	}
}
