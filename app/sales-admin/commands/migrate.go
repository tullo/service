package commands

import (
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4/database/postgres"
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

	var conf postgres.Config
	conf.StatementTimeout = 10 * time.Second
	driver, err := postgres.WithInstance(db.DB, &conf)
	if err != nil {
		return errors.Wrap(err, "migration driver construction")
	}

	if err := schema.Migrate(driver); err != nil {
		return errors.Wrap(err, "migrate database")
	}

	fmt.Println("migrations complete")

	return nil
}
