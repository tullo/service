package commands

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/tullo/service/business/data/user"
	"github.com/tullo/service/foundation/database"
)

// Users retrieves all users from the database.
func Users(traceID string, log *log.Logger, cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, err := user.Query(ctx, traceID, log, db)
	if err != nil {
		return errors.Wrap(err, "retrieve users")
	}

	return json.NewEncoder(os.Stdout).Encode(users)
}
