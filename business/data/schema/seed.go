// +build go1.16

package schema

import (
	"context"
	_ "embed" // go1.16 content embedding

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tullo/service/foundation/database"
)

// seeds is a string containing all of the queries needed to get the db seeded
// to a useful state for development.
//
// Note that database servers besides PostgreSQL may not support running
// multiple queries as part of the same execution so this single large constant
// may need to be broken up.
//
//go:embed sql/seed/data.sql
var seeds string

// Seed runs the set of seed-data queries against db. The queries are ran in a
// transaction and rolled back if any fail.
func Seed(ctx context.Context, db *sqlx.DB) error {
	if err := database.StatusCheck(ctx, db); err != nil {
		return errors.Wrap(err, "database status check")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(seeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// DeleteAll runs the set of drop-table queries against the database. The
// queries are run in a transaction and rolled back if any fail.
func DeleteAll(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(deleteAll); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// deleteAll is used to clean the database between tests.
const deleteAll = `
DELETE FROM sales;
DELETE FROM products;
DELETE FROM users;`
