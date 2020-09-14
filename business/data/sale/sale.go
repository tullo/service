package sale

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tullo/service/foundation/database"
	"go.opentelemetry.io/otel/api/trace"
)

// AddSale records a sales transaction for a single Product.
func AddSale(ctx context.Context, traceID string, log *log.Logger, db *sqlx.DB, ns NewSale, productID string, now time.Time) (*Sale, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.sale.add")
	defer span.End()

	s := Sale{
		ID:          uuid.New().String(),
		ProductID:   productID,
		Quantity:    ns.Quantity,
		Paid:        ns.Paid,
		DateCreated: now,
	}

	const q = `INSERT INTO sales
		(sale_id, product_id, quantity, paid, date_created)
		VALUES ($1, $2, $3, $4, $5)`

	log.Printf("%s : %s : query : %s", traceID, "sale.AddSale", database.Log(q,
		s.ID, s.ProductID, s.Quantity, s.Paid, s.DateCreated))

	_, err := db.ExecContext(ctx, q, s.ID, s.ProductID, s.Quantity, s.Paid, s.DateCreated)
	if err != nil {
		return nil, errors.Wrap(err, "inserting sale")
	}

	return &s, nil
}

// List gets all Sales from the database.
func List(ctx context.Context, traceID string, log *log.Logger, db *sqlx.DB, productID string) ([]Sale, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.sale.list")
	defer span.End()

	sales := []Sale{}

	const q = `SELECT * FROM sales WHERE product_id = $1`

	log.Printf("%s : %s : query : %s", traceID, "sale.List", database.Log(q, productID))

	if err := db.SelectContext(ctx, &sales, q, productID); err != nil {
		return nil, errors.Wrap(err, "selecting sales")
	}

	return sales, nil
}
