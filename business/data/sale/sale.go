package sale

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tullo/service/foundation/database"
	"go.opentelemetry.io/otel/trace"
)

// Sale manages the set of API's for sales access.
type Sale struct {
	log *log.Logger
	db  *sqlx.DB
}

// New constructs a Product for api access.
func New(log *log.Logger, db *sqlx.DB) Sale {
	return Sale{
		log: log,
		db:  db,
	}
}

// AddSale records a sales transaction for a single Product.
func (s Sale) AddSale(ctx context.Context, traceID string, ns NewSale, productID string, now time.Time) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.sale.add")
	defer span.End()

	sale := Info{
		ID:          uuid.New().String(),
		ProductID:   productID,
		Quantity:    ns.Quantity,
		Paid:        ns.Paid,
		DateCreated: now,
	}

	const q = `INSERT INTO sales
		(sale_id, product_id, quantity, paid, date_created)
		VALUES ($1, $2, $3, $4, $5)`

	s.log.Printf("%s : %s : query : %s", traceID, "sale.AddSale", database.Log(q,
		sale.ID, sale.ProductID, sale.Quantity, sale.Paid, sale.DateCreated))

	_, err := s.db.ExecContext(ctx, q, sale.ID, sale.ProductID, sale.Quantity, sale.Paid, sale.DateCreated)
	if err != nil {
		return Info{}, errors.Wrap(err, "inserting sale")
	}

	return sale, nil
}

// List gets all Sales from the database.
func (s Sale) List(ctx context.Context, traceID string, productID string) ([]Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.sale.list")
	defer span.End()

	sales := []Info{}

	const q = `SELECT * FROM sales WHERE product_id = $1`

	s.log.Printf("%s : %s : query : %s", traceID, "sale.List", database.Log(q, productID))

	if err := s.db.SelectContext(ctx, &sales, q, productID); err != nil {
		return nil, errors.Wrap(err, "selecting sales")
	}

	return sales, nil
}
