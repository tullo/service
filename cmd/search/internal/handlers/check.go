package handlers

import (
	"context"
	"net/http"

	"github.com/tullo/service/internal/platform/web"
	"go.opencensus.io/trace"
)

// Check provides support for orchestration health checks.
type Check struct {
	build string
}

// Health validates the service is healthy and ready to accept requests.
func (c *Check) Health(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Check.Health")
	defer span.End()

	health := struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	}{
		Status:  "ok",
		Version: c.build,
	}

	return web.Respond(ctx, w, health, http.StatusOK)
}
