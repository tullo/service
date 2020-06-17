package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/tullo/service/internal/mid"
	"github.com/tullo/service/internal/platform/auth" // Import is removed in final PR
	"github.com/tullo/service/internal/platform/web"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, db *sqlx.DB, authenticator *auth.Authenticator) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register health check endpoint. This route is not authenticated.
	check := check{
		build: build,
		db:    db,
	}
	app.Handle(http.MethodGet, "/v1/health", check.Health)

	// Register user management and authentication endpoints.
	u := user{
		db:            db,
		authenticator: authenticator,
	}
	app.Handle(http.MethodGet, "/v1/users", u.List, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodPost, "/v1/users", u.Create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/v1/users/{id}", u.Retrieve, mid.Authenticate(authenticator))
	app.Handle(http.MethodPut, "/v1/users/{id}", u.Update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodDelete, "/v1/users/{id}", u.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))

	// This route is not authenticated
	app.Handle(http.MethodGet, "/v1/users/token", u.Token)

	// Register product and sale endpoints.
	p := product{
		db: db,
	}
	app.Handle(http.MethodGet, "/v1/products", p.List, mid.Authenticate(authenticator))
	app.Handle(http.MethodPost, "/v1/products", p.Create, mid.Authenticate(authenticator))
	app.Handle(http.MethodGet, "/v1/products/{id}", p.Retrieve, mid.Authenticate(authenticator))
	app.Handle(http.MethodPut, "/v1/products/{id}", p.Update, mid.Authenticate(authenticator))
	app.Handle(http.MethodDelete, "/v1/products/{id}", p.Delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))

	app.Handle(http.MethodPost, "/v1/products/{id}/sales", p.AddSale, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/v1/products/{id}/sales", p.ListSales, mid.Authenticate(authenticator))

	return app
}
