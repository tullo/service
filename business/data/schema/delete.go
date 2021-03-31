package schema

import (
	"github.com/jmoiron/sqlx"
)

// DeleteAll runs the set of drop-table queries against the database. The
// queries are run in a transaction and rolled back if any fail.
func DeleteAll(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(deleteSQL); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}
