// Package database provides support for access the database.
package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.opentelemetry.io/otel/trace"
)

type DB struct {
	*pgxpool.Pool
}

// Config is the required properties to use the database.
type Config struct {
	User         string
	Password     string
	Host         string
	Name         string
	DisableTLS   bool
	MaxIdleConns int
	MaxOpenConns int
}

// Connect establishes a database connection based on the configuration.
func Connect(ctx context.Context, cfg Config) (*DB, error) {
	pool, err := pgxpool.Connect(ctx, ConnString(cfg))
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}
	db := DB{pool}

	return &db, nil
}

// ConnString translates the config to a db connection string.
func ConnString(cfg Config) string {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	// Query parameters.
	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	return u.String()
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *DB) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "foundation.database.statuscheck")
	defer span.End()

	// First check we can ping the database.
	var pingError error
	for attempts := 1; ; attempts++ {
		pingError = db.Ping(ctx)
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	// Make sure we didn't timeout or be cancelled.
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Run a simple query to determine connectivity. Running this query forces
	// a round trip to the database.
	const q = `SELECT true`
	var tmp bool
	return db.QueryRow(ctx, q).Scan(&tmp)
}
