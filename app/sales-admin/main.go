// This program performs administrative tasks for the sales-api service.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/tullo/conf"
	"github.com/tullo/service/app/sales-admin/commands"
	"github.com/tullo/service/foundation/config"
	"github.com/tullo/service/foundation/database"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {
	log := log.New(os.Stdout, "ADMIN : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	if err := run(log); err != nil {
		if errors.Cause(err) != commands.ErrHelp {
			log.Printf("error: %s", err)
		}
		os.Exit(1)
	}
}

func run(log *log.Logger) error {

	// =========================================================================
	// Configuration

	var cfg = config.NewCmdConfig(build, "copyright information here")
	if err := config.Parse(&cfg, config.SalesPrefix, os.Args[1:]); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			usage, err := config.Usage(&cfg, config.SalesPrefix)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		}
		if errors.Is(err, conf.ErrVersionWanted) {
			version, err := config.VersionString(&cfg, config.SalesPrefix)
			if err != nil {
				return errors.Wrap(err, "generating config version")
			}
			fmt.Println(version)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	// =========================================================================
	// Commands

	dbConfig := database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	}

	traceID := "00000000-0000-0000-0000-000000000000"

	switch cfg.Args.Num(0) {
	case "migrate":
		if err := commands.Migrate(dbConfig); err != nil {
			return errors.Wrap(err, "migrating database")
		}

	case "seed":
		if err := commands.Seed(dbConfig); err != nil {
			return errors.Wrap(err, "seeding database")
		}

	case "useradd":
		name := cfg.Args.Num(1)
		email := cfg.Args.Num(2)
		password := cfg.Args.Num(3)
		if err := commands.UserAdd(traceID, log, dbConfig, name, email, password); err != nil {
			return errors.Wrap(err, "adding user")
		}

	case "users":
		pageNumber := cfg.Args.Num(1)
		rowsPerPage := cfg.Args.Num(2)
		if err := commands.Users(traceID, log, dbConfig, pageNumber, rowsPerPage); err != nil {
			return errors.Wrap(err, "getting users")
		}

	case "keygen":
		if err := commands.KeyGen(); err != nil {
			return errors.Wrap(err, "key generation")
		}

	case "tokengen":
		userID := cfg.Args.Num(1)
		privateKeyFile := cfg.Args.Num(2)
		algorithm := cfg.Args.Num(3)
		if err := commands.TokenGen(traceID, log, dbConfig, userID, privateKeyFile, algorithm); err != nil {
			return errors.Wrap(err, "generating token")
		}

	default:
		fmt.Println("migrate: create the schema in the database")
		fmt.Println("seed: add data to the database")
		fmt.Println("useradd: add a new user to the database")
		fmt.Println("users: get a list of users from the database")
		fmt.Println("keygen: generate a set of private/public key files")
		fmt.Println("tokengen: generate a JWT for a user with claims")
		fmt.Println("provide a command to get more help.")
		return commands.ErrHelp
	}

	return nil
}
