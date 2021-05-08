package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data/schema"
	"github.com/tullo/service/business/data/user"
	"github.com/tullo/service/foundation/database"
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

// ContainerSpec provides configuration for a docker container to run.
type ContainerSpec struct {
	Repository string
	Tag        string
	Port       string
	Args       []string
}

type Container struct {
	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func NewContainer(pool, repository, tag string, env []string) (*Container, error) {
	p, err := dockertest.NewPool(pool)
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	hostConfig := func(hc *docker.HostConfig) {
		hc.AutoRemove = true // Auto remove stopped container.
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	}
	r, err := p.RunWithOptions(
		&dockertest.RunOptions{Repository: repository, Tag: tag, Env: env},
		hostConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("could not start docker container %w", err)
	}

	return &Container{
		pool:     p,
		resource: r,
	}, nil
}

func (c *Container) TailLogs(ctx context.Context, w io.Writer, follow bool) error {
	opts := docker.LogsOptions{
		Context: ctx,

		Stderr:      true,
		Stdout:      true,
		Follow:      follow,
		Timestamps:  true,
		RawTerminal: true,

		Container: c.resource.Container.ID,

		OutputStream: w,
	}

	return c.pool.Client.Logs(opts)
}

// Remove container and linked volumes from docker.
func removeContainer(t *testing.T, c *Container) {
	if err := c.pool.Purge(c.resource); err != nil {
		t.Error("Could not purge container:", err)
	}
}

func connect(c *Container, cfg database.Config) (*database.DB, error) {
	var db *database.DB
	// Connect using exponential backoff-retry.
	if err := c.pool.Retry(func() error {
		var (
			err error
			ctx = context.Background()
		)
		db, err = database.Connect(ctx, cfg)
		if err != nil {
			return err
		}
		return db.Ping(ctx)
	}); err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	return db, nil
}

func containerLog(t *testing.T, c *Container) {
	var buf bytes.Buffer
	c.TailLogs(context.Background(), &buf, false)
	t.Log(buf.String())
}

// NewUnit creates a test database inside a Docker container. It creates the
// required table structure but the database is otherwise empty. It returns
// the database to use as well as a function to call at the end of the test.
func NewUnit(t *testing.T, ctr ContainerSpec) (*log.Logger, *database.DB, func()) {
	log := log.New(os.Stdout, "TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	c, err := NewContainer("", ctr.Repository, ctr.Tag, ctr.Args)
	if err != nil {
		t.Fatal(err)
	}

	h := bytes.NewBufferString(c.resource.GetBoundIP(ctr.Port))
	fmt.Fprintf(h, ":%s", c.resource.GetPort(ctr.Port))
	cfg := database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       h.String(),
		Name:       "postgres",
		DisableTLS: true,
	}

	db, err := connect(c, cfg)
	if err != nil {
		containerLog(t, c)
		removeContainer(t, c)
		t.Fatalf("Opening database connection: %v", err)
	}

	if err := schema.Migrate(database.ConnString(cfg)); err != nil {
		containerLog(t, c)
		removeContainer(t, c)
		t.Fatalf("Migration error: %s", err)
	}

	// teardown is the function that should be invoked when
	// the caller is done with the database.
	teardown := func() {
		t.Helper()
		db.Close()
		removeContainer(t, c)
	}

	return log, db, teardown
}

// Test owns state for running and shutting down tests.
type Test struct {
	Auth     *auth.Auth
	DB       *database.DB
	KID      string
	Log      *log.Logger
	Teardown func()
	TraceID  string
	t        *testing.T
}

// NewIntegration creates a database, seeds it, constructs an authenticator.
func NewIntegration(t *testing.T, ctr ContainerSpec) *Test {

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
