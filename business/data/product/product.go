// Package product contains product related CRUD functionality.
package product

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data"
	"github.com/tullo/service/foundation/database"
	"go.opentelemetry.io/otel/trace"
)

// Product manages the set of API's for product access.
type Product struct {
	log *log.Logger
	db  *sqlx.DB
}

// New constructs a Product for api access.
func New(log *log.Logger, db *sqlx.DB) Product {
	return Product{
		log: log,
		db:  db,
	}
}

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func (p Product) Create(ctx context.Context, traceID string, claims auth.Claims, np NewProduct, now time.Time) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.create")
	defer span.End()

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
		(:product_id, :user_id, :name, :cost, :quantity, :date_created, :date_updated)`

	p.log.Printf("%s: %s: %s", traceID, "product.Create",
		database.Log(q, prd),
	)

	if _, err := p.db.NamedExecContext(ctx, q, prd); err != nil {
		return Info{}, errors.Wrap(err, "inserting product")
	}

	return prd, nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func (p Product) Update(ctx context.Context, traceID string, claims auth.Claims, productID string, up UpdateProduct, now time.Time) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.update")
	defer span.End()

	prd, err := p.QueryByID(ctx, traceID, productID)
	if err != nil {
		return err
	}

	if !claims.Authorized(auth.RoleAdmin) { // If you are not an admin
		if prd.UserID != claims.Subject { // and looking to retrieve someone elses product.
			return data.ErrForbidden
		}
	}

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
		"name" = :name,
		"cost" = :cost,
		"quantity" = :quantity,
		"date_updated" = :date_updated
	WHERE
		product_id = :product_id`

	p.log.Printf("%s: %s: %s", traceID, "product.Update",
		database.Log(q, prd),
	)

	if _, err = p.db.NamedExecContext(ctx, q, prd); err != nil {
		return errors.Wrap(err, "updating product")
	}

	return nil
}

// Delete removes the product identified by a given ID.
func (p Product) Delete(ctx context.Context, traceID string, claims auth.Claims, productID string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.delete")
	defer span.End()

	if _, err := uuid.Parse(productID); err != nil {
		return data.ErrInvalidID
	}

	// If you are not an admin.
	if !claims.Authorized(auth.RoleAdmin) {
		return data.ErrForbidden
	}

	filter := struct {
		ProductID string `db:"product_id"`
	}{
		ProductID: productID,
	}

	const q = `
	DELETE FROM
		products
	WHERE
		product_id = :product_id`

	p.log.Printf("%s: %s: %s", traceID, "product.Delete",
		database.Log(q, filter),
	)

	if _, err := p.db.NamedExecContext(ctx, q, filter); err != nil {
		return errors.Wrapf(err, "deleting product %s", productID)
	}

	return nil
}

// Query gets all Products from the database.
func (p Product) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.query")
	defer span.End()

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
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	page := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	p.log.Printf("%s: %s: %s", traceID, "product.Query",
		database.Log(q, page),
	)

	ns, err := p.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "prepare named context")
	}
	defer ns.Close()

	products := make([]Info, 0, page.RowsPerPage)
	if err = ns.SelectContext(ctx, &products, page); err != nil {
		return nil, errors.Wrap(err, "query products")
	}

	return products, nil
}

// QueryByID finds the product identified by a given ID.
func (p Product) QueryByID(ctx context.Context, traceID string, productID string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.querybyid")
	defer span.End()

	if _, err := uuid.Parse(productID); err != nil {
		return Info{}, data.ErrInvalidID
	}

	filter := struct {
		ProductID string `db:"product_id"`
	}{
		ProductID: productID,
	}

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
		p.product_id = :product_id
	GROUP BY
		p.product_id`

	p.log.Printf("%s: %s: %s", traceID, "product.QueryByID",
		database.Log(q, filter),
	)

	ns, err := p.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return Info{}, errors.Wrap(err, "prepare named context")
	}
	defer ns.Close()

	var prd Info
	if err := ns.GetContext(ctx, &prd, filter); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, data.ErrNotFound
		}

		return Info{}, errors.Wrapf(err, "selecting product %q", productID)
	}

	return prd, nil
}
