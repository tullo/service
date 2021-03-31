// +build go1.16

package schema

import (
	"context"
	_ "embed" // go1.16 content embedding

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tullo/service/foundation/database"
)

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

	if _, err := tx.Exec(seedSQL); err != nil {
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
