package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/tullo/service/business/data/schema"
	"github.com/tullo/service/foundation/database"
)

// Seed loads test data into the database.
func Seed(cfg database.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "connect to database")
	}
	defer db.Close()

	if err := schema.Seed(ctx, db); err != nil {
		return errors.Wrap(err, "seed database")
	}

	fmt.Println("seed data complete")
	return nil
}
