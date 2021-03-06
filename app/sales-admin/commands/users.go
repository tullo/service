package commands

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/tullo/service/business/data/user"
	"github.com/tullo/service/foundation/database"
)

// Users retrieves all users from the database.
func Users(traceID string, log *log.Logger, cfg database.Config, pageNumber string, rowsPerPage string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	page, err := strconv.Atoi(pageNumber)
	if err != nil {
		return errors.Wrap(err, "converting page number")
	}

	rows, err := strconv.Atoi(rowsPerPage)
	if err != nil {
		return errors.Wrap(err, "converting rows per page")
	}

	u := user.NewStore(log, db)
	users, err := u.Query(ctx, traceID, page, rows)
	if err != nil {
		return errors.Wrap(err, "retrieve users")
	}

	return json.NewEncoder(os.Stdout).Encode(users)
}
