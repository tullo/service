package mid

import (
	"context"
	"expvar"
	"net/http"
	"runtime"
	"strings"

	"github.com/tullo/service/foundation/web"
	"go.opentelemetry.io/otel"
)

// m contains the global program counters for the application.
var m = struct {
	gr  *expvar.Int
	req *expvar.Int
	err *expvar.Int
}{
	gr:  expvar.NewInt("goroutines"),
	req: expvar.NewInt("requests"),
	err: expvar.NewInt("errors"),
}

// Metrics updates program counters.
func Metrics() web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Wrap this handler around the next one provided.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			ctx, span := otel.Tracer(name).Start(ctx, "business.mid.metrics")
			defer span.End()

			// Don't count anything on /debug routes towards metrics.
			// Call the next handler to continue processing.
			if strings.HasPrefix(r.URL.Path, "/debug") {
				return handler(ctx, w, r)
			}

			// Call the next handler.
			err := handler(ctx, w, r)

			// Increment the request counter.
			m.req.Add(1)

			// Update the count for the number of active goroutines every 100 requests.
			if m.req.Value()%100 == 0 {
				m.gr.Set(int64(runtime.NumGoroutine()))
			}

			// Increment the errors counter if an error occurred on this request.
			if err != nil {
				m.err.Add(1)
			}

			// Return the error so it can be handled further up the chain.
			return err
		}

		return h
	}

	return m
}
