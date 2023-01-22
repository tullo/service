package mid

import (
	"context"
	"net/http"

	"github.com/tullo/service/foundation/web"
	"go.opentelemetry.io/otel"
)

// Cors sets the response headers needed for Cross-Origin Resource Sharing.
//
// Should be applied at the route-level unless every single route needs these
// headers.
//
// mid.Cors(corsOrigin)
func Cors(origin string) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := otel.Tracer(name).Start(ctx, "business.mid.cors")
			defer span.End()

			// Set the CORS headers on the response.
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, DELETE, GET, POST, PUT")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Accept-Encoding, Authorization, Content-Type, Content-Length, X-CSRF-Token")

			// Call the next handler.
			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
