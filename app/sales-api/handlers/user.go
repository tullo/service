package handlers

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data/user"
	"github.com/tullo/service/foundation/web"
	"go.opentelemetry.io/otel/api/global"
)

// User represents the User API method handler set.
type userHandlers struct {
	db   *sqlx.DB
	auth *auth.Auth

	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// Query returns all the existing users in the system.
func (h *userHandlers) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := global.Tracer("service").Start(ctx, "handlers.user.query")
	defer span.End()

	users, err := user.Query(ctx, h.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, users, http.StatusOK)
}

// QueryByID returns the specified user from the system.
func (h *userHandlers) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := global.Tracer("service").Start(ctx, "handlers.user.queryByID")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	id := web.Param(r, "id")
	usr, err := user.QueryByID(ctx, claims, h.db, id)
	if err != nil {
		switch err {
		case user.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case user.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case user.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "Id: %s", id)
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

// Create inserts a new user into the system.
func (h *userHandlers) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := global.Tracer("service").Start(ctx, "handlers.user.create")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var nu user.NewUser
	if err := web.Decode(r, &nu); err != nil {
		return errors.Wrap(err, "")
	}

	usr, err := user.Create(ctx, h.db, nu, v.Now)
	if err != nil {
		return errors.Wrapf(err, "User: %+v", &usr)
	}

	return web.Respond(ctx, w, usr, http.StatusCreated)
}

// Update updates the specified user in the system.
func (h *userHandlers) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := global.Tracer("service").Start(ctx, "handlers.user.update")
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
		return errors.Wrap(err, "")
	}

	id := web.Param(r, "id")
	err := user.Update(ctx, claims, h.db, id, upd, v.Now)
	if err != nil {
		switch err {
		case user.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case user.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case user.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s  User: %+v", id, &upd)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Delete removes the specified user from the system.
func (h *userHandlers) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := global.Tracer("service").Start(ctx, "handlers.user.delete")
	defer span.End()

	id := web.Param(r, "id")
	err := user.Delete(ctx, h.db, id)
	if err != nil {
		switch err {
		case user.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case user.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case user.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "Id: %s", id)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Token handles a request to authenticate a user. It expects a request using
// Basic Auth with a user's email and password. It responds with a JWT.
func (h *userHandlers) token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := global.Tracer("service").Start(ctx, "handlers.user.token")
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

	claims, err := user.Authenticate(ctx, h.db, v.Now, email, pass)
	if err != nil {
		switch err {
		case user.ErrAuthenticationFailure:
			return web.NewRequestError(err, http.StatusUnauthorized)
		default:
			return errors.Wrap(err, "authenticating")
		}
	}

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = h.auth.GenerateToken(claims)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}
