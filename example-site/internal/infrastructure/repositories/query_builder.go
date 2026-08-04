package repositories

import (
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

var (
	psql = goqu.Dialect("postgres")
)

func init() { // nolint: gochecknoinits
	goqu.SetDefaultPrepared(true)
}
