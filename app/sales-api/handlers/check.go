package handlers

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/tullo/service/foundation/database"
	"github.com/tullo/service/foundation/web"
	"go.opentelemetry.io/otel/api/global"
)

// Check provides support for orchestration health checks.
type check struct {
	build string
	db    *sqlx.DB

	// ADD OTHER STATE LIKE THE LOGGER IF NEEDED.
}

// Health validates the service is healthy and ready to accept requests.
func (h *check) Health(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.check.health")
	defer span.End()

	health := struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	}{
		Version: h.build,
	}

	if r.Header.Get("X-Probe") == "LivenessProbe" {
		health.Status = "ok"
		return web.Respond(ctx, w, health, http.StatusOK)
	}

	// Check if the database is ready.
	if err := database.StatusCheck(ctx, h.db); err != nil {

		// If the database is not ready we will tell the client and use a 500
		// status. Do not respond by just returning an error because further up in
		// the call stack will interpret that as an unhandled error.
		health.Status = "db not ready"
		return web.Respond(ctx, w, health, http.StatusInternalServerError)
	}

	health.Status = "ok"
	return web.Respond(ctx, w, health, http.StatusOK)
}
