package commands

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/tullo/service/internal/data"
	"github.com/tullo/service/internal/platform/database"
)

// Users retrieves all users from the database.
func Users(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, err := data.Retrieve.User.List(ctx, db)
	if err != nil {
		return errors.Wrap(err, "retrieve users")
	}

	return json.NewEncoder(os.Stdout).Encode(users)
}
