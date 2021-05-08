// Package schema contains the database schema, migrations and seeding data.

// +build go1.16

package schema

import (
	"embed" // go1.16 content embedding
)

var (
	// deleteSQL is used to clean the database between tests.
	//
	//go:embed sql/delete.sql
	deleteSQL string

	// Migrations contains the migrations needed to construct
	// the database schema. Migration file pairs (up/down)
	// should never be removed from this directory once they
	// have been run in production.
	//
	//go:embed migrations
	migrations embed.FS

	// seedSQL is a string containing all the queries needed
	// to get the db seeded to a useful state for development.
	//
	// Note that database servers besides PostgreSQL may not
	// support running multiple queries as part of the same
	// execution so this single large constant may need to
	// be broken up.
	//
	//go:embed sql/seed.sql
	seedSQL string
)
