package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/tullo/service/cmd/search/internal/views"
	"github.com/tullo/service/internal/platform/web"
	"github.com/tullo/service/internal/product"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
)

// Search provides support for orchestration searches.
type Search struct {
	log *log.Logger
	url string
}

// NewSearch constructs a Search for a given set of feeds.
func NewSearch(url string, log *log.Logger) *Search {
	return &Search{
		log: log,
		url: url,
	}
}

// Query performs a search against the datastore.
func (s *Search) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Search.Query")
	defer span.End()

	// Create a new request.
	req, err := http.NewRequest(http.MethodGet, s.url, nil)
	if err != nil {
		return err
	}

	// Create child span for backend call
	ctx, childspan := trace.StartSpan(ctx, s.url)
	defer childspan.End()

	// Create a context with a timeout of 1 second.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// Bind the new context into the request.
	req = req.WithContext(ctx)

	// Add span context to backend call
	format := &tracecontext.HTTPFormat{}
	format.SpanContextToRequest(childspan.SpanContext(), req)

	// Make the web call and return any error. Do will handle the
	// context level timeout.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// Close the response body on the return.
	defer resp.Body.Close()

	// Decode the results.
	var products []product.Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return err
	}

	// Render the results as HTML.
	markup, err := views.Render(products, s.url)
	if err != nil {
		return err
	}

	web.RespondHTML(ctx, w, markup, http.StatusOK)
	return nil
}
