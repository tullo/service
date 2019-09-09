package handlers

import (
	"context"
	"log"
	"net/http"
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

	// var results []views.Result
	if r.Method == "POST" {
		// var searchers []search.Searcher
		// results = search.Submit(ctx, s.log, options.Term, options.First, searchers)
	}

	// markup, err := views.Render(fv, results)
	// if err != nil {
	// 	return err
	// }

	// web.RespondHTML(ctx, w, markup, http.StatusOK)
	return nil
}
