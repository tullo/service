// Package product contains product related CRUD functionality.
package product

import (
	"context"
	"log"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data"
	"github.com/tullo/service/foundation/database"
	"go.opentelemetry.io/otel/trace"
)

// Store manages the set of API's for product access.
type Store struct {
	log *log.Logger
	db  *database.DB
}

// NewStore constructs a Store for api access.
func NewStore(log *log.Logger, db *database.DB) Store {
	return Store{
		log: log,
		db:  db,
	}
}

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func (s Store) Create(ctx context.Context, traceID string, claims auth.Claims, np NewProduct, now time.Time) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.create")
	defer span.End()

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return Info{}, errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

	prd := Info{
		ID:          uuid.New().String(),
		Name:        np.Name,
		Cost:        np.Cost,
		Quantity:    np.Quantity,
		UserID:      claims.Subject,
		DateCreated: now.UTC(),
		DateUpdated: now.UTC(),
	}

	const q = `
	INSERT INTO products
		(product_id, user_id, name, cost, quantity, date_created, date_updated)
	VALUES
		($1, $2, $3, $4, $5, $6, $7)`

	if _, err := conn.Exec(ctx, q, prd.ID, prd.UserID, prd.Name, prd.Cost, prd.Quantity, prd.DateCreated, prd.DateUpdated); err != nil {
		return Info{}, errors.Wrap(err, "inserting product")
	}

	return prd, nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func (s Store) Update(ctx context.Context, traceID string, claims auth.Claims, productID string, up UpdateProduct, now time.Time) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.update")
	defer span.End()

	prd, err := s.QueryByID(ctx, traceID, productID)
	if err != nil {
		return err
	}

	if !claims.Authorized(auth.RoleAdmin) { // If you are not an admin
		if prd.UserID != claims.Subject { // and looking to retrieve someone elses product.
			return data.ErrForbidden
		}
	}

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

	if up.Name != nil {
		prd.Name = *up.Name
	}
	if up.Cost != nil {
		prd.Cost = *up.Cost
	}
	if up.Quantity != nil {
		prd.Quantity = *up.Quantity
	}
	prd.DateUpdated = now

	const q = `
	UPDATE
		products
	SET
		"name" = $2,
		"cost" = $3,
		"quantity" = $4,
		"date_updated" = $5
	WHERE
		product_id = $1`

	if _, err = conn.Exec(ctx, q, productID, prd.Name, prd.Cost, prd.Quantity, prd.DateUpdated); err != nil {
		return errors.Wrap(err, "updating product")
	}

	return nil
}

// Delete removes the product identified by a given ID.
func (s Store) Delete(ctx context.Context, traceID string, claims auth.Claims, productID string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.delete")
	defer span.End()

	if _, err := uuid.Parse(productID); err != nil {
		return data.ErrInvalidID
	}

	// If you are not an admin.
	if !claims.Authorized(auth.RoleAdmin) {
		return data.ErrForbidden
	}

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

	const q = `
	DELETE FROM
		products
	WHERE
		product_id = $1`

	if _, err := conn.Exec(ctx, q, productID); err != nil {
		return errors.Wrapf(err, "deleting product %s", productID)
	}

	return nil
}

// Query gets all Products from the database.
func (s Store) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.query")
	defer span.End()

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

	const q = `
	SELECT
		p.*,
		COALESCE(SUM(s.quantity) ,0) AS sold,
		COALESCE(SUM(s.paid), 0) AS revenue
	FROM
		products AS p
	LEFT JOIN
		sales AS s ON p.product_id = s.product_id
	GROUP BY
		p.product_id
	ORDER BY
		user_id
	OFFSET $1 ROWS FETCH NEXT $2 ROWS ONLY`

	page := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	products := make([]Info, 0, page.RowsPerPage)
	if err := pgxscan.Select(ctx, conn, &products, q, page.Offset, page.RowsPerPage); err != nil {
		return nil, errors.Wrap(err, "query products")
	}

	return products, nil
}

// QueryByID finds the product identified by a given ID.
func (s Store) QueryByID(ctx context.Context, traceID string, productID string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.querybyid")
	defer span.End()

	if _, err := uuid.Parse(productID); err != nil {
		return Info{}, data.ErrInvalidID
	}

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return Info{}, errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

	const q = `
	SELECT
		p.*,
		COALESCE(SUM(s.quantity), 0) AS sold,
		COALESCE(SUM(s.paid), 0) AS revenue
	FROM
		products AS p
	LEFT JOIN
		sales AS s ON p.product_id = s.product_id
	WHERE
		p.product_id = $1
	GROUP BY
		p.product_id`

	var prd Info
	if err := pgxscan.Get(ctx, conn, &prd, q, productID); err != nil {
		if pgxscan.NotFound(err) {
			return Info{}, data.ErrNotFound
		}

		return Info{}, errors.Wrapf(err, "selecting product %q", productID)
	}

	return prd, nil
}
