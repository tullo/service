package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data"
	"github.com/tullo/service/business/data/product"
	"github.com/tullo/service/business/data/sale"
	"github.com/tullo/service/foundation/web"
	"go.opentelemetry.io/otel/api/global"
)

// Product represents the Product API method handler set.
type productHandlers struct {
	db *sqlx.DB
}

// List gets all existing products in the system.
func (h *productHandlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := global.Tracer("service").Start(ctx, "handlers.product.list")
	defer span.End()

	products, err := product.List(ctx, h.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, products, http.StatusOK)
}

// Retrieve returns the specified product from the system.
func (h *productHandlers) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := global.Tracer("service").Start(ctx, "handlers.product.retrieve")
	defer span.End()

	id := web.Param(r, "id")
	prod, err := product.One(ctx, h.db, id)
	if err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case data.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", id)
		}
	}

	return web.Respond(ctx, w, prod, http.StatusOK)
}

// Create decodes the body of a request to create a new product. The full
// product with populatd fields is sent back in the response.
func (h *productHandlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := global.Tracer("service").Start(ctx, "handlers.product.create")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return web.NewShutdownError("claims missing from context")
	}

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var np product.NewProduct
	if err := web.Decode(r, &np); err != nil {
		return errors.Wrap(err, "decoding new product")
	}

	prod, err := product.Create(ctx, h.db, claims, np, v.Now)
	if err != nil {
		return errors.Wrapf(err, "creating new product: %+v", np)
	}

	return web.Respond(ctx, w, prod, http.StatusCreated)
}

// Update decodes the body of a request to update an existing product. The ID
// of the product is part of the request URL.
func (h *productHandlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := global.Tracer("service").Start(ctx, "handlers.product.update")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return web.NewShutdownError("claims missing from context")
	}

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var up product.UpdateProduct
	if err := web.Decode(r, &up); err != nil {
		return errors.Wrap(err, "")
	}

	id := web.Param(r, "id")
	if err := product.Update(ctx, h.db, claims, id, up, v.Now); err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case data.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case data.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "updating product %q: %+v", id, up)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Delete removes a single product identified by an ID in the request URL.
func (h *productHandlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := global.Tracer("service").Start(ctx, "handlers.product.delete")
	defer span.End()

	id := web.Param(r, "id")
	if err := product.Delete(ctx, h.db, id); err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "Id: %s", id)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// AddSale creates a new Sale for a particular product. It looks for a JSON
// object in the request body. The full model is returned to the caller.
func (h *productHandlers) AddSale(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := global.Tracer("service").Start(ctx, "handlers.product.addSale")
	defer span.End()

	var ns sale.NewSale
	if err := web.Decode(r, &ns); err != nil {
		return errors.Wrap(err, "decoding new sale")
	}

	id := web.Param(r, "id")
	sale, err := sale.AddSale(r.Context(), h.db, ns, id, time.Now())
	if err != nil {
		return errors.Wrap(err, "adding new sale")
	}

	return web.Respond(ctx, w, sale, http.StatusCreated)
}

// ListSales gets all sales for a particular product.
func (h *productHandlers) ListSales(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := global.Tracer("service").Start(ctx, "handlers.product.listSales")
	defer span.End()

	id := web.Param(r, "id")
	list, err := sale.List(r.Context(), h.db, id)
	if err != nil {
		return errors.Wrap(err, "getting sales list")
	}

	return web.Respond(ctx, w, list, http.StatusOK)
}
