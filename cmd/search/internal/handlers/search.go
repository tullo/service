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
	"go.opencensus.io/trace"
)

// Search provides support for orchestration searches.
type Search struct {
	log *log.Logger
}

// NewSearch constructs a Search for a given set of feeds.
func NewSearch(log *log.Logger) *Search {
	return &Search{
		log: log,
	}
}

// Query performs a search against the datastore.
func (s *Search) Query(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Search.Query")
	defer span.End()

	// Create a new request.
	req, err := http.NewRequest("GET", "http://sales-api:3000/v1/products", nil)
	if err != nil {
		return err
	}

	bearer := "Bearer " + "eyJhbGciOiJSUzI1NiIsImtpZCI6IjEiLCJ0eXAiOiJKV1QifQ.eyJyb2xlcyI6WyJBRE1JTiIsIlVTRVIiXSwiZXhwIjoxNTY4MzM4NTQ5LCJpYXQiOjE1NjgzMzQ5NDksInN1YiI6IjVjZjM3MjY2LTM0NzMtNDAwNi05ODRmLTkzMjUxMjI2NzhiNyJ9.X3BkW54YN-ecc_xhfGeQi6tqLLn43sn7ejzRYFFMQ8T31-dEfA13XUMA7bST-ADzusVn8FZORiOfhHBKGtwFCMGN9ArkethUnNTEtdX72WimCkohLPBAoVg3DTKgNihqo4nVNzoB7B27CqlthyYBJs6fHBJIEsq-L4TsKL2a_97HabSR1gow5_5yQ7V48gA2Nn6V_ECqn76A5MEwq_DOXgTapLDoIStrr-X-Se2DVnSQfxxG3PmDCzhJqrhWFTNvSWi-ShL7zh7SmJNqWwMXQn5K2tAX_7wayl6ABhtX5UANs6oxRUsGC6UQjFwOwLkBe0GV_6pnWb7oXPz0rOsS3w"
	req.Header.Set("Authorization", bearer)

	// Create a context with a timeout of 1 second.
	ctx, cancel := context.WithTimeout(req.Context(), time.Second)
	defer cancel()

	// Bind the new context into the request.
	req = req.WithContext(ctx)

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
	markup, err := views.Render(products)
	if err != nil {
		return err
	}

	web.RespondHTML(ctx, w, markup, http.StatusOK)
	return nil
}
