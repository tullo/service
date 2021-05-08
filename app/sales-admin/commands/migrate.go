package commands

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/tullo/service/business/data/schema"
	"github.com/tullo/service/foundation/database"
)

// ErrHelp provides context that help was given.
var ErrHelp = errors.New("provided help")

// Migrate creates the schema in the database.
func Migrate(cfg database.Config) error {
	if err := schema.Migrate(database.ConnString(cfg)); err != nil {
		return errors.Wrap(err, "migrate database")
	}

	fmt.Println("migrations complete")

	return nil
}
