package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data/user"
	"github.com/tullo/service/foundation/database"
)

// UserAdd adds new users into the database.
func UserAdd(cfg database.Config, email, password string) error {
	if email == "" || password == "" {
		fmt.Println("help: useradd <email> <password>")
		return ErrHelp
	}

	db, err := database.Open(cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nu := user.NewUser{
		Email:           email,
		Password:        password,
		PasswordConfirm: password,
		Roles:           []string{auth.RoleAdmin, auth.RoleUser},
	}
	u, err := user.Create(ctx, db, nu, time.Now())
	if err != nil {
		return errors.Wrap(err, "create user")
	}

	fmt.Println("user id:", u.ID)
	return nil
}
