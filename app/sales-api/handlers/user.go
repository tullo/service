package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data"
	"github.com/tullo/service/business/data/user"
	"github.com/tullo/service/foundation/web"
	"go.opentelemetry.io/otel"
)

// userGroup represents the User API method handler set.
type userGroup struct {
	user user.Store
	auth *auth.Auth
}

// Query returns all the existing users in the system.
func (ug userGroup) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := otel.Tracer(name).Start(ctx, "handlers.user.query")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	page := web.Param(r, "page")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("invalid page format: %s", page), http.StatusBadRequest)
	}
	rows := web.Param(r, "rows")
	rowsPerPage, err := strconv.Atoi(rows)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("invalid rows format: %s", rows), http.StatusBadRequest)
	}

	users, err := ug.user.Query(ctx, v.TraceID, pageNumber, rowsPerPage)
	if err != nil {
		return errors.Wrap(err, "unable to query for users")
	}

	return web.Respond(ctx, w, users, http.StatusOK)
}

// QueryByID returns the specified user from the system.
func (ug userGroup) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := otel.Tracer(name).Start(ctx, "handlers.user.queryByID")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	id := web.Param(r, "id")
	usr, err := ug.user.QueryByID(ctx, v.TraceID, claims, id)
	if err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case data.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case data.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s", id)
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

// Create inserts a new user into the system.
func (ug userGroup) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := otel.Tracer(name).Start(ctx, "handlers.user.create")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var nu user.NewUser
	if err := web.Decode(r, &nu); err != nil {
		return errors.Wrap(err, "decoding new user")
	}

	usr, err := ug.user.Create(ctx, v.TraceID, nu, v.Now)
	if err != nil {
		return errors.Wrapf(err, "User: %+v", &usr)
	}

	return web.Respond(ctx, w, usr, http.StatusCreated)
}

// Update updates the specified user in the system.
func (ug userGroup) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := otel.Tracer(name).Start(ctx, "handlers.user.update")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	var upd user.UpdateUser
	if err := web.Decode(r, &upd); err != nil {
		return errors.Wrap(err, "decoding updated user")
	}

	id := web.Param(r, "id")
	err := ug.user.Update(ctx, v.TraceID, claims, id, upd, v.Now)
	if err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case data.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case data.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s  User: %+v", id, &upd)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Delete removes the specified user from the system.
func (ug userGroup) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := otel.Tracer(name).Start(ctx, "handlers.user.delete")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	id := web.Param(r, "id")
	err := ug.user.Delete(ctx, v.TraceID, claims, id)
	if err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case data.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case data.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s", id)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Token handles a request to authenticate a user. It expects a request using
// Basic Auth with a user's email and password. It responds with a JWT.
func (ug userGroup) token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := otel.Tracer(name).Start(ctx, "handlers.user.token")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := ug.user.Authenticate(ctx, v.TraceID, v.Now, email, pass)
	if err != nil {
		switch err {
		case data.ErrAuthenticationFailure:
			return web.NewRequestError(err, http.StatusUnauthorized)
		default:
			return errors.Wrap(err, "authenticating")
		}
	}

	kid := web.Param(r, "kid")
	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = ug.auth.GenerateToken(kid, claims)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}
