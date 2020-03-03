package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tullo/service/internal/platform/auth"
	"github.com/tullo/service/internal/platform/web"
	"github.com/tullo/service/internal/product"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
)

// Product represents the Product API method handler set.
type Product struct {
	db *sqlx.DB

	// ADD OTHER STATE LIKE THE LOGGER IF NEEDED.
}

// List gets all existing products in the system.
func (p *Product) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := trace.StartSpan(ctx, "handlers.Product.List")

	// Get span context from incoming request
	HTTPFormat := &tracecontext.HTTPFormat{}
	if spanContext, ok := HTTPFormat.SpanContextFromRequest(r); ok {
		ctx, span = trace.StartSpanWithRemoteParent(ctx, "handlers.Product.List", spanContext)
	}
	defer span.End()

	products, err := product.List(ctx, p.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, products, http.StatusOK)
}

// Retrieve returns the specified product from the system.
func (p *Product) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := trace.StartSpan(ctx, "handlers.Product.Retrieve")

	// Get span context from incoming request
	HTTPFormat := &tracecontext.HTTPFormat{}
	if spanContext, ok := HTTPFormat.SpanContextFromRequest(r); ok {
		ctx, span = trace.StartSpanWithRemoteParent(ctx, "handlers.Product.Retrieve", spanContext)
	}
	defer span.End()

	id := chi.URLParam(r, "id")
	prod, err := product.Retrieve(ctx, p.db, id)
	if err != nil {
		switch err {
		case product.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case product.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", id)
		}
	}

	return web.Respond(ctx, w, prod, http.StatusOK)
}

// Create decodes the body of a request to create a new product. The full
// product with generated fields is sent back in the response.
func (p *Product) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := trace.StartSpan(ctx, "handlers.Product.Create")

	// Get span context from incoming request
	HTTPFormat := &tracecontext.HTTPFormat{}
	if spanContext, ok := HTTPFormat.SpanContextFromRequest(r); ok {
		ctx, span = trace.StartSpanWithRemoteParent(ctx, "handlers.Product.Create", spanContext)
	}
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

	prod, err := product.Create(ctx, p.db, claims, np, v.Now)
	if err != nil {
		return errors.Wrapf(err, "creating new product: %+v", np)
	}

	return web.Respond(ctx, w, prod, http.StatusCreated)
}

// Update decodes the body of a request to update an existing product. The ID
// of the product is part of the request URL.
func (p *Product) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := trace.StartSpan(ctx, "handlers.Product.Update")

	// Get span context from incoming request
	HTTPFormat := &tracecontext.HTTPFormat{}
	if spanContext, ok := HTTPFormat.SpanContextFromRequest(r); ok {
		ctx, span = trace.StartSpanWithRemoteParent(ctx, "handlers.Product.Update", spanContext)
	}
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

	id := chi.URLParam(r, "id")
	if err := product.Update(ctx, p.db, claims, id, up, v.Now); err != nil {
		switch err {
		case product.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case product.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case product.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "updating product %q: %+v", id, up)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Delete removes a single product identified by an ID in the request URL.
func (p *Product) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := trace.StartSpan(ctx, "handlers.Product.Delete")

	// Get span context from incoming request
	HTTPFormat := &tracecontext.HTTPFormat{}
	if spanContext, ok := HTTPFormat.SpanContextFromRequest(r); ok {
		ctx, span = trace.StartSpanWithRemoteParent(ctx, "handlers.Product.Delete", spanContext)
	}
	defer span.End()

	id := chi.URLParam(r, "id")
	if err := product.Delete(ctx, p.db, id); err != nil {
		switch err {
		case product.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "Id: %s", id)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// AddSale creates a new Sale for a particular product. It looks for a JSON
// object in the request body. The full model is returned to the caller.
func (p *Product) AddSale(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := trace.StartSpan(ctx, "handlers.Product.AddSale")

	// Get span context from incoming request
	HTTPFormat := &tracecontext.HTTPFormat{}
	if spanContext, ok := HTTPFormat.SpanContextFromRequest(r); ok {
		ctx, span = trace.StartSpanWithRemoteParent(ctx, "handlers.Product.AddSale", spanContext)
	}
	defer span.End()

	var ns product.NewSale
	if err := web.Decode(r, &ns); err != nil {
		return errors.Wrap(err, "decoding new sale")
	}

	id := chi.URLParam(r, "id")
	sale, err := product.AddSale(r.Context(), p.db, ns, id, time.Now())
	if err != nil {
		return errors.Wrap(err, "adding new sale")
	}

	return web.Respond(ctx, w, sale, http.StatusCreated)
}

// ListSales gets all sales for a particular product.
func (p *Product) ListSales(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	ctx, span := trace.StartSpan(ctx, "handlers.Product.ListSales")

	// Get span context from incoming request
	HTTPFormat := &tracecontext.HTTPFormat{}
	if spanContext, ok := HTTPFormat.SpanContextFromRequest(r); ok {
		ctx, span = trace.StartSpanWithRemoteParent(ctx, "handlers.Product.ListSales", spanContext)
	}
	defer span.End()

	id := chi.URLParam(r, "id")
	list, err := product.ListSales(r.Context(), p.db, id)
	if err != nil {
		return errors.Wrap(err, "getting sales list")
	}

	return web.Respond(ctx, w, list, http.StatusOK)
}
