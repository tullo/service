package mid

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/foundation/web"
	"go.opentelemetry.io/otel/api/trace"
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(a *auth.Auth) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Wrap this handler around the next one provided.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.mid.authenticate")
			defer span.End()

			parts := strings.Split(r.Header.Get("Authorization"), " ")
			// Expecting header format `Bearer <token>`.
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.New("expected authorization header format: Bearer <token>")
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			// Validate the token is signed by us.
			claims, err := a.ValidateToken(parts[1])
			if err != nil {
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			// Add claims to the context so they can be retrieved later.
			ctx = context.WithValue(ctx, auth.Key, claims)

			// Call the next handler.
			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// Authorize validates that an authenticated user has at least one role from a
// specified list. This method constructs the actual function that is used.
func Authorize(roles ...string) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.mid.hasrole")
			defer span.End()

			claims, ok := ctx.Value(auth.Key).(auth.Claims)
			if !ok {
				return errors.New("claims missing from context")
			}

			if !claims.Authorized(roles...) {
				return web.NewRequestError(
					fmt.Errorf("you are not authorized for that action: claims: %v exp: %v", claims.Roles, roles),
					http.StatusForbidden,
				)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
