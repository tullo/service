package schema

import (
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
)

// Migrate attempts to bring the schema for the database up to date with the
// migrations defined in this package.
func Migrate(driver database.Driver) error {
	src, err := httpfs.New(http.FS(migrations), "migrations")
	if err != nil {
		return err
	}

	mig, err := migrate.NewWithInstance("httpfs", src, "postgres", driver)
	if err != nil {
		return err
	}

	if err = mig.Up(); err != migrate.ErrNoChange {
		return err
	}

	return nil
}
