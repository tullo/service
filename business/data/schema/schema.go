// +build go1.16

package schema

import (
	_ "embed"
)

var (
	// schemaSQL contains the queries needed to construct
	// the database schema. Entries should never be removed
	// from this file once they have been run in production.
	//
	//go:embed sql/schema.sql
	schemaSQL string

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
