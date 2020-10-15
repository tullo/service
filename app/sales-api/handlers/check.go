package handlers

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/tullo/service/foundation/database"
	"github.com/tullo/service/foundation/web"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/codes"
)

// Check provides support for orchestration health checks.
type checkGroup struct {
	build string
	db    *sqlx.DB
	log   *log.Logger
	// ADD OTHER STATE LIKE THE LOGGER IF NEEDED.
}

// readiness checks if the database is ready and if not will return a 500 status.
// Do not respond by just returning an error because further up in the call
// stack it will interpret that as a non-trusted error.
func (cg checkGroup) readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.check.readiness")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	status := "ok"
	statusCode := http.StatusOK
	if err := database.StatusCheck(ctx, v.TraceID, cg.log, cg.db); err != nil {
		status = "db not ready"
		statusCode = http.StatusInternalServerError
		span.SetStatus(codes.Error, web.CheckErr(err))
		span.AddEvent(ctx, "Database is not ready!")
	}

	readiness := struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	}{
		Version: cg.build,
		Status:  status,
	}

	return web.Respond(ctx, w, readiness, statusCode)
}

// liveness returns simple status info if the service is alive. If the
// app is deployed to a Kubernetes cluster, it will also return pod, node, and
// namespace details via the Downward API. The Kubernetes environment variables
// need to be set within your Pod/Deployment manifest.
func (cg checkGroup) liveness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.check.liveness")
	defer span.End()

	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	info := struct {
		Status    string `json:"status,omitempty"`
		Build     string `json:"build,omitempty"`
		Host      string `json:"host,omitempty"`
		Pod       string `json:"pod,omitempty"`
		PodIP     string `json:"podIP,omitempty"`
		Node      string `json:"node,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	}{
		Status:    "up",
		Build:     cg.build,
		Host:      host,
		Pod:       os.Getenv("KUBERNETES_PODNAME"),
		PodIP:     os.Getenv("KUBERNETES_NAMESPACE_POD_IP"),
		Node:      os.Getenv("KUBERNETES_NODENAME"),
		Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
	}

	return web.Respond(ctx, w, info, http.StatusOK)
}
