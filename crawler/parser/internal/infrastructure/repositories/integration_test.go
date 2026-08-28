//go:build integration

package repositories

import (
	"os"
	"testing"

	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/infrastructure/repositories/testutil"
)

func TestMain(m *testing.M) {
	if err := testutil.Setup(); err != nil {
		println("failed to setup test database: " + err.Error())
		os.Exit(1)
	}
	defer testutil.Teardown()
	os.Exit(m.Run())
}
