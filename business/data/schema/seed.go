package schema

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tullo/service/foundation/database"
)

// Seed runs the set of seed-data queries against db. The queries are ran in a
// transaction and rolled back if any fail.
func Seed(ctx context.Context, db *database.DB) error {
	if err := database.StatusCheck(ctx, db); err != nil {
		return errors.Wrap(err, "database status check")
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, seedSQL); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return err
		}
		return err
	}

	return tx.Commit(ctx)
}
