package commands

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/tullo/service/internal/data"
	"github.com/tullo/service/internal/platform/database"
)

// Seed loads test data into the database.
func Seed(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	if err := data.Seed(db); err != nil {
		return errors.Wrap(err, "seed database")
	}

	fmt.Println("seed data complete")
	return nil
}
