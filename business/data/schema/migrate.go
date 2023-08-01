package schema

import (
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/pkg/errors"
)

// Migrate attempts to bring the schema for the database up to date with the
// migrations defined in this package.
func Migrate(connString string) error {
	var c cockroachdb.CockroachDb
	driver, err := c.Open(connString + "&x-statement-timeout=10000") // 10 seconds
	if err != nil {
		return errors.Wrap(err, "migration driver construction")
	}

	src, err := httpfs.New(http.FS(migrations), "migrations")
	if err != nil {
		return errors.Wrap(err, "create migrate source driver")
	}

	mig, err := migrate.NewWithInstance("httpfs", src, connString, driver)
	if err != nil {
		return errors.Wrap(err, "create migrate instance")
	}

	if err = mig.Up(); err != migrate.ErrNoChange {
		return err
	}

	return nil
}
