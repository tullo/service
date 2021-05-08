package schema

import (
	"context"

	"github.com/tullo/service/foundation/database"
)

// DeleteAll runs delete-from-table queries against the database.
// The queries are run in a transaction and rolled back if any fail.
func DeleteAll(db *database.DB) error {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, deleteSQL); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return err
		}
		return err
	}

	return tx.Commit(ctx)
}
