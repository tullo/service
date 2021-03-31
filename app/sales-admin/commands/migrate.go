package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/tullo/service/business/data/schema"
	"github.com/tullo/service/foundation/database"
)

// ErrHelp provides context that help was given.
var ErrHelp = errors.New("provided help")

// Migrate creates the schema in the database.
func Migrate(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return errors.Wrap(err, "connect to database")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Migrate(ctx, db); err != nil {
		return errors.Wrap(err, "migrate database")
	}

	fmt.Println("migrations complete")
	return nil
}
