// +build go1.16

package schema_test

import (
	_ "embed"
	"testing"
)

//go:embed sql/seed/data.sql
var seeds string

func TestEmbedSeedData(t *testing.T) {
	if len(seeds) == 0 {
		t.Error("embedding of seed data failed")
	}
	t.Logf("\n%s", seeds)
}
