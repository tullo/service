package tests

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data/schema"
	"github.com/tullo/service/business/data/user"
	"github.com/tullo/service/foundation/database"
	"github.com/tullo/service/foundation/docker"
	"github.com/tullo/service/foundation/keystore"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

// Configuration for running tests.
var (
	// IDs from the seed data for admin@example.com and user@example.com.
	AdminID = "5cf37266-3473-4006-984f-9325122678b7"
	UserID  = "45b5fbd3-755f-4379-8f07-a58d4a30fa2f"
)

// Container provides configuration for a docker container to run.
type Container struct {
	Image string
	Port  string
	Args  []string
}

// NewUnit creates a test database inside a Docker container. It creates the
// required table structure but the database is otherwise empty. It returns
// the database to use as well as a function to call at the end of the test.
func NewUnit(t *testing.T, ctr Container) (*log.Logger, *sqlx.DB, func()) {

	// Start a DB container instance with dgraph running.
	c := docker.StartContainer(t, ctr.Image, ctr.Port, ctr.Args...)

	cfg := database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.Host,
		Name:       "postgres",
		DisableTLS: true,
	}
	db, err := database.Open(cfg)

	if err != nil {
		t.Fatalf("Opening database connection: %v", err)
	}

	t.Log("Waiting for database to be ready ...")

	// Wait for the database to be ready.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Migrate(ctx, db); err != nil {
		docker.DumpContainerLogs(t, c.ID)
		docker.StopContainer(t, c.ID)
		t.Fatalf("Migration error: %s", err)
	}

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown := func() {
		t.Helper()
		db.Close()
		docker.StopContainer(t, c.ID)
	}

	log := log.New(os.Stdout, "TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	return log, db, teardown
}

// Test owns state for running and shutting down tests.
type Test struct {
	Auth     *auth.Auth
	DB       *sqlx.DB
	KID      string
	Log      *log.Logger
	Teardown func()
	TraceID  string
	t        *testing.T
}

// NewIntegration creates a database, seeds it, constructs an authenticator.
func NewIntegration(t *testing.T, ctr Container) *Test {

	// Initialize and seed database. Store the cleanup function call later.
	log, db, teardown := NewUnit(t, ctr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Seed(ctx, db); err != nil {
		t.Fatal(err)
	}

	// Create RSA keys to enable authentication in our service.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	// The corresponding public key ID.
	keyID := "4754d86b-7a6d-4df5-9c65-224741361492"

	// Build an authenticator using this private key and id for the key store.
	keyPair := map[string]*rsa.PrivateKey{keyID: privateKey}
	keyStore := keystore.NewMap(keyPair)
	auth, err := auth.New("RS256", keyStore)
	if err != nil {
		t.Fatal(err)
	}

	test := Test{
		Auth:     auth,
		DB:       db,
		KID:      keyID,
		Log:      log,
		Teardown: teardown,
		TraceID:  "00000000-0000-0000-0000-000000000000",
		t:        t,
	}

	return &test
}

// Token generates an authenticated token for a user.
func (test *Test) Token(email, pass string) string {
	test.t.Log("Generating token for test ...")
	u := user.NewStore(test.Log, test.DB)
	claims, err := u.Authenticate(context.Background(), test.TraceID, time.Now(), email, pass)
	if err != nil {
		test.t.Fatal(err)
	}

	token, err := test.Auth.GenerateToken(test.KID, claims)
	if err != nil {
		test.t.Fatal(err)
	}

	return token
}

// StringPointer is a helper to get a *string from a string. It is in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func StringPointer(s string) *string {
	return &s
}

// IntPointer is a helper to get a *int from a int. It is in the tests package
// because we normally don't want to deal with pointers to basic types but it's
// useful in some tests.
func IntPointer(i int) *int {
	return &i
}
