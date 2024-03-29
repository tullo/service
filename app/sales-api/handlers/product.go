package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data"
	"github.com/tullo/service/business/data/product"
	"github.com/tullo/service/business/data/sale"
	"github.com/tullo/service/foundation/web"
	"go.opentelemetry.io/otel"
)

// productGroup represents the Product API method handler set.
type productGroup struct {
	product product.Store
	sale    sale.Store
}

// Query gets all existing products in the system.
func (pg productGroup) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := otel.Tracer(name).Start(ctx, "handlers.product.query")
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

	products, err := pg.product.Query(ctx, v.TraceID, pageNumber, rowsPerPage)
	if err != nil {
		return errors.Wrap(err, "unable to query products")
	}

	return web.Respond(ctx, w, products, http.StatusOK)
}

// QueryByID returns the specified product from the system.
func (pg productGroup) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := otel.Tracer(name).Start(ctx, "handlers.product.queryByID")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	id := web.Param(r, "id")
	prod, err := pg.product.QueryByID(ctx, v.TraceID, id)
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
func (pg productGroup) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := otel.Tracer(name).Start(ctx, "handlers.product.create")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return web.NewShutdownError("claims missing from context")
	}

	var np product.NewProduct
	if err := web.Decode(r, &np); err != nil {
		return errors.Wrap(err, "unable to decode payload")
	}

	prod, err := pg.product.Create(ctx, v.TraceID, claims, np, v.Now)
	if err != nil {
		return errors.Wrapf(err, "creating new product: %+v", np)
	}

	return web.Respond(ctx, w, prod, http.StatusCreated)
}

// Update decodes the body of a request to update an existing product. The ID
// of the product is part of the request URL.
func (pg productGroup) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := otel.Tracer(name).Start(ctx, "handlers.product.update")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return web.NewShutdownError("claims missing from context")
	}

	var up product.UpdateProduct
	if err := web.Decode(r, &up); err != nil {
		return errors.Wrap(err, "decoding updated product")
	}

	id := web.Param(r, "id")
	if err := pg.product.Update(ctx, v.TraceID, claims, id, up, v.Now); err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case data.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case data.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %q Product: %+v", id, up)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Delete removes a single product identified by an ID in the request URL.
func (pg productGroup) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := otel.Tracer(name).Start(ctx, "handlers.product.delete")
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
	if err := pg.product.Delete(ctx, v.TraceID, claims, id); err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "ID: %s", id)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// AddSale creates a new Sale for a particular product. It looks for a JSON
// object in the request body. The full model is returned to the caller.
func (pg productGroup) addSale(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := otel.Tracer(name).Start(ctx, "handlers.product.addSale")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var ns sale.NewSale
	if err := web.Decode(r, &ns); err != nil {
		return errors.Wrap(err, "decoding new sale")
	}

	id := web.Param(r, "id")
	sale, err := pg.sale.AddSale(r.Context(), v.TraceID, ns, id, time.Now())
	if err != nil {
		return errors.Wrap(err, "adding new sale")
	}

	return web.Respond(ctx, w, sale, http.StatusCreated)
}

// QuerySales gets all sales for a particular product.
func (pg productGroup) querySales(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := otel.Tracer(name).Start(ctx, "handlers.product.querySales")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	id := web.Param(r, "id")
	list, err := pg.sale.List(r.Context(), v.TraceID, id)
	if err != nil {
		return errors.Wrap(err, "getting sales list")
	}

	return web.Respond(ctx, w, list, http.StatusOK)
}
