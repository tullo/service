package commands

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data/user"
	"github.com/tullo/service/foundation/database"
)

// UserAdd adds new users into the database.
func UserAdd(traceID string, log *log.Logger, cfg database.Config, name, email, password string) error {

	if name == "" || email == "" || password == "" {
		fmt.Println("help: useradd <name> <email> <password>")
		return ErrHelp
	}

	db, err := database.Open(cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	u := user.NewStore(log, db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nu := user.NewUser{
		Name:            name,
		Email:           email,
		Password:        password,
		PasswordConfirm: password,
		Roles:           []string{auth.RoleAdmin, auth.RoleUser},
	}

	usr, err := u.Create(ctx, traceID, nu, time.Now())
	if err != nil {
		return errors.Wrap(err, "create user")
	}

	fmt.Println("user id:", usr.ID)
	return nil
}
