package seller

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTryCreateAdvert_WhenSuccessful_ThenSendsUpsertRequest(t *testing.T) {
	var (
		mu      sync.Mutex
		records []RecordedRequest
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		body, _ := io.ReadAll(r.Body)
		records = append(records, RecordedRequest{
			Method: r.Method,
			URL:    r.URL.String(),
			Body:   body,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"title":"Test","description":"Test desc","price":100,"pubDate":"2024-01-01T00:00:00Z"}`))
	}))
	defer srv.Close()

	seller, err := New(srv.URL, Config{Region: "test-region"}, zap.NewNop())
	require.NoError(t, err)

	err = seller.TryCreateAdvert(context.Background())
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, records, 1)
	assert.Equal(t, http.MethodPut, records[0].Method)
	assert.Contains(t, records[0].URL, "/adverts/test-region/adverts/")

	var adv map[string]any
	err = json.Unmarshal(records[0].Body, &adv)
	require.NoError(t, err)
	assert.NotEmpty(t, adv["title"])
	assert.NotEmpty(t, adv["price"])
}

func TestTryDeleteAdverts_WhenSuccessful_ThenSendsSearchAndDeleteRequests(t *testing.T) {
	var (
		mu      sync.Mutex
		records []RecordedRequest
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		body, _ := io.ReadAll(r.Body)
		records = append(records, RecordedRequest{
			Method: r.Method,
			URL:    r.URL.String(),
			Body:   body,
		})

		switch r.Method {
		case http.MethodGet:
			switch r.URL.Path {
			case "/adverts/test-region/adverts/adv1":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"title":"Test","description":"Test desc","price":100,"pubDate":"1970-01-01T00:00:00Z"}`))
			case "/adverts/test-region/search":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"page":1,"total":1,"nextPage":null,"adverts":[{"id":"adv1","title":"Test Ad","url":"/adverts/test-region/adverts/adv1"}]}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}

		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))
	defer srv.Close()

	seller, err := New(srv.URL, Config{Region: "test-region", DeleteJitter: time.Second, DeleteBatchSize: 1000}, zap.NewNop())
	require.NoError(t, err)

	err = seller.TryDeleteAdverts(context.Background())
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, records, 3)
	assert.Equal(t, http.MethodGet, records[0].Method)
	assert.Contains(t, records[0].URL, "/adverts/test-region/search")

	assert.Equal(t, http.MethodGet, records[1].Method)
	assert.Contains(t, records[1].URL, "/adverts/test-region/adverts/adv1")

	assert.Equal(t, http.MethodDelete, records[2].Method)
	assert.Contains(t, records[2].URL, "/adverts/test-region/adverts/adv1")
}

type RecordedRequest struct {
	Method string
	URL    string
	Body   []byte
}
