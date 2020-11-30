package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data/product"
	"github.com/tullo/service/business/data/sale"
	"github.com/tullo/service/business/data/user"
	"github.com/tullo/service/business/mid"
	"github.com/tullo/service/foundation/web"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, db *sqlx.DB, a *auth.Auth) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register debug check endpoints. This routes are not authenticated.
	cg := checkGroup{
		build: build,
		db:    db,
		log:   log,
	}

	app.HandleDebug(http.MethodGet, "/readiness", cg.readiness)
	app.HandleDebug(http.MethodGet, "/liveness", cg.liveness)

	// Register user management and authentication endpoints.
	ug := userGroup{
		user: user.New(log, db),
		auth: a,
	}

	app.Handle(http.MethodGet, "/v1/users/{page}/{rows}", ug.query, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodPost, "/v1/users", ug.create, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/v1/users/{id}", ug.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/v1/users/{id}", ug.update, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodDelete, "/v1/users/{id}", ug.delete, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	// This route is not authenticated
	app.Handle(http.MethodGet, "/v1/users/token", ug.token)

	// Register product and sale endpoints.
	pg := productGroup{
		product: product.New(log, db),
		sale:    sale.New(log, db),
	}
	app.Handle(http.MethodGet, "/v1/products/{page}/{rows}", pg.query, mid.Authenticate(a))
	app.Handle(http.MethodPost, "/v1/products", pg.create, mid.Authenticate(a))
	app.Handle(http.MethodGet, "/v1/products/{id}", pg.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/v1/products/{id}", pg.update, mid.Authenticate(a))
	app.Handle(http.MethodDelete, "/v1/products/{id}", pg.delete, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))

	app.Handle(http.MethodPost, "/v1/products/{id}/sales", pg.addSale, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/v1/products/{id}/sales", pg.querySales, mid.Authenticate(a))

	return app
}
