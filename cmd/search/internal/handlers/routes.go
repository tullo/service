package handlers

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/tullo/service/internal/mid"
	"github.com/tullo/service/internal/platform/web"
)

// API returns a handler for a set of routes.
func API(build, url string, shutdown chan os.Signal, log *log.Logger) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	pwd, _ := os.Getwd()
	if "/app" != pwd {
		pwd += "/cmd/search"
	}
	static := pwd + "/static"
	// Because our static directory is set as the root of the FileSystem,
	// we need to strip off the /static/ prefix from the request path
	// before searching the FileSystem for the given file.
	fs := http.FileServer(http.Dir(static))
	sp := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
		return nil
	}
	app.Handle(http.MethodGet, "/static/*", sp, nil)

	// Register health check endpoint. This route is not authenticated.
	check := Check{
		build: build,
	}
	app.Handle(http.MethodGet, "/health", check.Health)

	// Register health check endpoint. This route is not authenticated.
	search := NewSearch(url, log)
	app.Handle(http.MethodGet, "/search", search.Query)
	app.Handle(http.MethodPost, "/search", search.Query)

	return app
}
