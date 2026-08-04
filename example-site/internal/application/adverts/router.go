//go:generate go tool ogen --config .ogen.yml --target gen --package gen --clean ../../../docs/openapi/openapi.yml
package adverts

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/application/adverts/gen"
	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain"
)

func NewRouter(
	logger *zap.Logger,
	advertsService *domain.AdvertsCRUDService,
) (http.Handler, error) {
	handler := NewHandler(logger, advertsService)
	srv, err := gen.NewServer(
		handler,
	)
	fn := func(w http.ResponseWriter, r *http.Request) {
		route, ok := srv.FindRoute(r.Method, r.URL.Path)
		if ok {
			r.Pattern = route.PathPattern() // for net/http middlewares compatibility
		}
		srv.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn), err
}
