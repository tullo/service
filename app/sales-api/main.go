package main

import (
	"context"
	"crypto/rsa"
	"expvar" // Register the expvar handlers
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof" // Register the pprof handlers
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tullo/conf"
	"github.com/tullo/service/app/sales-api/handlers"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/foundation/config"
	"github.com/tullo/service/foundation/database"
	"github.com/tullo/service/foundation/tracer"
)

/*
Need to figure out timeouts for http service.
You might want to reset your DB_HOST env var during test tear down.
Service should start even without a DB running yet.
symbols in profiles: https://github.com/golang/go/issues/23376 / https://github.com/google/pprof/pull/366
*/

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

type deps struct {
	auth    *auth.Auth
	db      *sqlx.DB
	cfg     *config.AppConfig
	log     *log.Logger
	srverr  chan error
	srvdown chan os.Signal
}

func main() {
	log := log.New(os.Stdout, "SALES : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	if err := run(log); err != nil {
		log.Println("main: error:", err)
		os.Exit(1)
	}
}

func run(log *log.Logger) error {

	// Print the build version for our logs.
	log.Printf("main: Application initializing : version %q", build)
	defer log.Println("main: Completed")

	// Expose the build version under /debug/vars.
	expvar.NewString("build").Set(build)

	var err error

	// =========================================================================
	// Configuration

	var cfg = config.NewAppConfig(build, "copyright information here")
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

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating output for config")
	}
	log.Printf("main: Config :\n%v\n", out)

	// =========================================================================
	// Initialize authentication support

	var auth *auth.Auth
	if auth, err = initAuthSupport(log, &cfg); err != nil {
		return errors.Wrap(err, "init auth support")
	}

	// =========================================================================
	// Start Database Support

	log.Println("main: Initializing database support")

	db, err := database.Open(database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	})
	if err != nil {
		return errors.Wrap(err, "connecting to db")
	}
	defer func() {
		log.Printf("main: Database Stopping : %s", cfg.DB.Host)
		db.Close()
	}()

	// =========================================================================
	// Start Tracing Support

	log.Println("main: Initializing zipkin tracing support")

	tr := tracer.Config{
		ServiceName: cfg.Zipkin.ServiceName,
		ReporterURI: cfg.Zipkin.ReporterURI,
		Probability: cfg.Zipkin.Probability,
	}
	if err = tracer.Init(log, &tr); err != nil {
		return errors.Wrap(err, "starting tracer")
	}

	// =========================================================================
	// Start Debug Service

	startDebugService(log, &cfg)

	// =========================================================================
	// Start API Service

	log.Println("main: Initializing API support")

	d := deps{
		auth:    auth,
		db:      db,
		cfg:     &cfg,
		log:     log,
		srverr:  nil,
		srvdown: nil,
	}
	api := initAPI(&d)

	// Start the service listening for requests.
	go func() {
		log.Printf("main: API listening on %s", api.Addr)
		d.srverr <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	timeout := &cfg.Web.ShutdownTimeout
	appShutdown(api, d.srvdown, d.srverr, timeout)

	return nil
}

// startDebugService launches a goroutine serving registered debug handlers.
//
// /debug/pprof - handler added to the default mux by importing the net/http/pprof package.
// /debug/vars - handler added to the default mux by importing the expvar package.
//
// Not concerned with shutting this down when the application is shutdown.
func startDebugService(log *log.Logger, cfg *config.AppConfig) {
	log.Println("main: Initializing debugging support")

	go func() {
		log.Printf("main: Debug Listening %s", cfg.Web.DebugHost)
		if err := http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux); err != nil {
			log.Printf("main: Debug Listener closed : %v", err)
		}
	}()
}

func initAuthSupport(log *log.Logger, cfg *config.AppConfig) (*auth.Auth, error) {
	log.Println("main: Initializing authentication support")

	privatePEM, err := ioutil.ReadFile(cfg.Auth.PrivateKeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading auth private key")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return nil, errors.Wrap(err, "parsing auth private key")
	}

	lookup := func(kid string) (*rsa.PublicKey, error) {
		switch kid {
		case cfg.Auth.KeyID:
			return &privateKey.PublicKey, nil
		}
		return nil, fmt.Errorf("no public key found for the specified kid: %s", kid)
	}
	auth, err := auth.New(cfg.Auth.Algorithm, lookup)
	if err != nil {
		return nil, errors.Wrap(err, "constructing authenticator")
	}
	auth.AddKey(cfg.Auth.KeyID, privateKey)

	return auth, nil
}

func initAPI(d *deps) *http.Server {
	log.Println("main: Initializing API support")

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	d.srverr = make(chan error, 1)

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	d.srvdown = make(chan os.Signal, 1)
	signal.Notify(d.srvdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:         d.cfg.Web.APIHost,
		Handler:      handlers.API(build, d.srvdown, d.log, d.db, d.auth),
		ReadTimeout:  d.cfg.Web.ReadTimeout,
		WriteTimeout: d.cfg.Web.WriteTimeout,
	}

	return &api
}

func appShutdown(api *http.Server, down chan os.Signal, srverr chan error, d *time.Duration) error {
	// Blocking main and waiting for shutdown.
	select {
	case err := <-srverr:
		return errors.Wrap(err, "server error")

	case sig := <-down:
		log.Printf("main: %v : Start shutdown", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), *d)
		defer cancel()

		// Asking listener to shutdown and shed load.
		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return errors.Wrap(err, "could not stop server gracefully")
		}
	}

	return nil
}
