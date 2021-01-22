// Package config provides configuration support.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/tullo/conf"
)

// SalesPrefix is use for sales related config parsing.
const SalesPrefix = "SALES"

// CmdConfig holds common configuration properties.
type CmdConfig struct {
	conf.Version
	Args conf.Args
	DB   struct {
		User       string `conf:"default:postgres"`
		Password   string `conf:"default:postgres,noprint"`
		Host       string `conf:"default:0.0.0.0"`
		Name       string `conf:"default:postgres"`
		DisableTLS bool   `conf:"default:false"`
	}
}

// AppConfig holds application configuration properties.
type AppConfig struct {
	conf.Version
	Web struct {
		APIHost         string        `conf:"default:0.0.0.0:3000"`
		DebugHost       string        `conf:"default:0.0.0.0:4000"`
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:5s"`
		ShutdownTimeout time.Duration `conf:"default:5s"`
	}
	DB struct {
		User       string `conf:"default:postgres"`
		Password   string `conf:"default:postgres,noprint"`
		Host       string `conf:"default:0.0.0.0"`
		Name       string `conf:"default:postgres"`
		DisableTLS bool   `conf:"default:false"`
	}
	Auth struct {
		KeyID          string `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"`
		PrivateKeyFile string `conf:"default:/service/private.pem"`
		Algorithm      string `conf:"default:RS256"`
	}
	Zipkin struct {
		ReporterURI string  `conf:"default:http://zipkin:9411/api/v2/spans"`
		ServiceName string  `conf:"default:sales-api"`
		Probability float64 `conf:"default:0.05"`
	}
}

// Parse parses configuration into the provided struct.
func Parse(cfg interface{}, prefix string) error {
	if err := conf.Parse(os.Args[1:], prefix, cfg); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage(prefix, cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
			version, err := conf.VersionString(prefix, cfg)
			if err != nil {
				return errors.Wrap(err, "generating config version")
			}
			fmt.Println(version)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	return nil
}
