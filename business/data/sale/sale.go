package sale

import (
	"context"
	"log"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tullo/service/foundation/database"
	"go.opentelemetry.io/otel"
)

const name = "sale"

// Store manages the set of API's for sales access.
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

// AddSale records a sales transaction for a single Product.
func (s Store) AddSale(ctx context.Context, traceID string, ns NewSale, productID string, now time.Time) (Info, error) {
	ctx, span := otel.Tracer(name).Start(ctx, "business.data.sale.add")
	defer span.End()

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return Info{}, errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

	sale := Info{
		ID:          uuid.New().String(),
		ProductID:   productID,
		Quantity:    ns.Quantity,
		Paid:        ns.Paid,
		DateCreated: now,
	}

	const q = `INSERT INTO sales (sale_id, product_id, quantity, paid, date_created)
		VALUES ($1, $2, $3, $4, $5)`

	_, err = conn.Exec(ctx, q, sale.ID, sale.ProductID, sale.Quantity, sale.Paid, sale.DateCreated)
	if err != nil {
		return Info{}, errors.Wrap(err, "inserting sale")
	}

	return sale, nil
}

// List gets all Sales from the database.
func (s Store) List(ctx context.Context, traceID string, productID string) ([]Info, error) {
	ctx, span := otel.Tracer(name).Start(ctx, "business.data.sale.list")
	defer span.End()

	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "acquire db connection")
	}
	defer conn.Release()

	const q = `SELECT * FROM sales WHERE product_id = $1`

	var sales []Info
	if err := pgxscan.Select(ctx, conn, &sales, q, productID); err != nil {
		return nil, errors.Wrap(err, "selecting sales")
	}

	return sales, nil
}
