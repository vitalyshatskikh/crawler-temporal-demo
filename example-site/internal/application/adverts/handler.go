package adverts

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/application/adverts/gen"
	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain"
)

var _ gen.Handler = (*Handler)(nil)

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
	advert, err := h.service.GetAdvert(ctx, domain.AdvertIdentity{Region: params.Region, ID: params.ID})
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
	result, err := h.service.SearchAdverts(
		ctx,
		domain.AdvertSearchParams{
			Region:     params.Region,
			PageSize:   params.Size.Value,
			PageNum:    params.Page.Value,
			OlderFirst: params.OlderFirst.Value,
		},
	)
	if err != nil {
		return nil, err
	}
	items := make([]gen.AdvertListItem, 0, len(result.Adverts))
	for _, adv := range result.Adverts {
		advURL := fmt.Sprintf("/adverts/%s/adverts/%s", url.PathEscape(adv.Region), url.PathEscape(adv.ID))
		items = append(items, gen.AdvertListItem{ID: adv.ID, Title: adv.Title, URL: advURL})
	}

	nextPage := gen.OptNilString{}
	if result.AdvertsTotal > result.PageNum*result.PageSize {
		next := fmt.Sprintf("/adverts/%s/search?page=%d&size=%d",
			url.PathEscape(params.Region), result.PageNum+1, params.Size.Value)
		nextPage = gen.NewOptNilString(next)
	}

	return &gen.SearchResponse{
		Page:     result.PageNum,
		Total:    result.AdvertsTotal,
		NextPage: nextPage,
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
	err := h.service.DeleteAdvert(ctx, domain.AdvertIdentity{Region: params.Region, ID: params.ID})
	if err != nil {
		return err
	}
	return nil
}

func (h Handler) NewError(_ context.Context, err error) *gen.ErrorStatusCode {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, domain.ErrValidation):
		return newErrorResponse(http.StatusBadRequest, "validation error",
			map[string]string{"cause": err.Error()})
	case errors.Is(err, domain.ErrNotFound):
		return newErrorResponse(http.StatusNotFound, "not found", nil)
	default:
		h.logger.Error("internal server error", zap.Error(err))
		return newErrorResponse(http.StatusInternalServerError, "internal server error", nil)
	}
}

func newErrorResponse(status int, msg string, detail map[string]string) *gen.ErrorStatusCode {
	resp := gen.Error{Error: msg}
	if detail != nil {
		resp.Detail = gen.OptErrorDetail{Set: true, Value: detail}
	}
	return &gen.ErrorStatusCode{StatusCode: status, Response: resp}
}
