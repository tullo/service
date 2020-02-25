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
	req, err := http.NewRequest("GET", s.url+"/v1/products", nil)
	if err != nil {
		return err
	}

	bearer := "Bearer " + "eyJhbGciOiJSUzI1NiIsImtpZCI6IjEiLCJ0eXAiOiJKV1QifQ.eyJyb2xlcyI6WyJBRE1JTiIsIlVTRVIiXSwiZXhwIjoxNTgyNjM4NTk5LCJpYXQiOjE1ODI2MzQ5OTksInN1YiI6IjVjZjM3MjY2LTM0NzMtNDAwNi05ODRmLTkzMjUxMjI2NzhiNyJ9.Rkp6MvYXOPIL04-lCACyGyIZqpP--XN59VqQSFktJDhe5WK5_wTDSpeBNeACNE1F6JRBQx30_CD3mMriF68MdqJy5Ui0YWl76stxUK1AnvHrGE9h0UTCswYyOCySX2o1alCPuzbtQGDI5OL4bfKtoAodlbbUVxP_UJRqo98xo3OPvRq7V3MK7yE-RDG0KdM1RAsYasw1O2uE3ESVEMRKXJIAFHBg843BK_Kv4m30WdILNfGjUX3tHgkBQM9pfWLOg4dGXY0nfIZZ-eseQAdUKw_jmuqt18TU5_jSjK7sqnhG93sHlJKHnnUqg8VywRrJwm0esOIPyZBmRuTUGYVaGg"
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
