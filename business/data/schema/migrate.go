// Package schema contains the database schema, migrations and seeding data.
package schema

import (
	"bufio"
	"context"
	_ "embed"
	"strconv"
	"strings"

	"github.com/ardanlabs/darwin"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tullo/service/foundation/database"
)

// Migrate attempts to bring the schema for db up to date with the migrations
// defined in this package.
func Migrate(ctx context.Context, db *sqlx.DB) error {
	if err := database.StatusCheck(ctx, db); err != nil {
		return errors.Wrap(err, "database status check")
	}

	driver, err := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})
	if err != nil {
		return errors.Wrap(err, "construct darwin driver")
	}

	d := darwin.New(driver, parseMigrations(schemaSQL))
	return d.Migrate()
}

func parseMigrations(s string) []darwin.Migration {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanLines)

	var migs []darwin.Migration
	var mig darwin.Migration
	var index int = -1
	for scanner.Scan() {
		v := strings.ToLower(scanner.Text())
		switch {
		// Migration version.
		case len(v) >= 5 && (v[:6] == "-- ver" || v[:5] == "--ver"):
			f, err := strconv.ParseFloat(strings.TrimSpace(v[11:]), 64)
			if err != nil {
				return nil
			}
			mig.Version = f
			migs = append(migs, mig)

			index++
			mig = darwin.Migration{}

		// Migration description.
		case len(v) >= 5 && (v[:6] == "-- des" || v[:5] == "--des"):
			migs[index].Description = strings.TrimSpace(v[15:])

		default:
			migs[index].Script += v + "\n"
		}
	}

	return migs
}
