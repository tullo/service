// Package config provides configuration support.
package config_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tullo/conf"
	"github.com/tullo/service/foundation/config"
)

func TestVersionString(t *testing.T) {
	type args struct {
		cfg    interface{}
		prefix string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"Cmd Empty",
			args{cfg: &config.CmdConfig{}, prefix: "TEST"},
			"",
			false,
		},
		{
			"Cmd V1",
			args{cfg: &config.CmdConfig{Version: conf.Version{Version: "v1.0.0"}}, prefix: "TEST"},
			"Version: v1.0.0",
			false,
		},
		{
			"Cmd V1 Desc",
			args{cfg: &config.CmdConfig{Version: conf.Version{Description: "Description of v1.0.0"}}, prefix: "TEST"},
			"Description of v1.0.0",
			false,
		},
		{
			"App V1",
			args{cfg: &config.AppConfig{Version: conf.Version{Version: "v1.0.0"}}, prefix: "TEST"},
			"Version: v1.0.0",
			false,
		},
		{
			"App V1 Desc",
			args{cfg: &config.AppConfig{Version: conf.Version{Description: "Description of v1.0.0"}}, prefix: "TEST"},
			"Description of v1.0.0",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.VersionString(tt.args.cfg, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("VersionString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("VersionString() got = %v, want %v", got, tt.want)
				t.Log(cmp.Diff(got, tt.want))
			}
		})
	}
}

var cmdConfigHelp string = `Usage: config.test [options] [arguments]

OPTIONS
  --db-user/$TEST_DB_USER                <string>  (default: admin)
  --db-password/$TEST_DB_PASSWORD        <string>  (default: postgres)
  --db-host/$TEST_DB_HOST                <string>  (default: 0.0.0.0:26257)
  --db-name/$TEST_DB_NAME                <string>  (default: postgres)
  --db-disable-tls/$TEST_DB_DISABLE_TLS  <bool>    (default: false)
  --help/-h                              
  display this help message
  --version/-v  
  display version information
`

var appConfigHelp string = `Usage: config.test [options] [arguments]

OPTIONS
  --web-api-host/$TEST_WEB_API_HOST                  <string>    (default: 0.0.0.0:3000)
  --web-debug-host/$TEST_WEB_DEBUG_HOST              <string>    (default: 0.0.0.0:4000)
  --web-read-timeout/$TEST_WEB_READ_TIMEOUT          <duration>  (default: 5s)
  --web-write-timeout/$TEST_WEB_WRITE_TIMEOUT        <duration>  (default: 5s)
  --web-shutdown-timeout/$TEST_WEB_SHUTDOWN_TIMEOUT  <duration>  (default: 5s)
  --db-user/$TEST_DB_USER                            <string>    (default: admin)
  --db-password/$TEST_DB_PASSWORD                    <string>    (default: postgres)
  --db-host/$TEST_DB_HOST                            <string>    (default: 0.0.0.0:26257)
  --db-name/$TEST_DB_NAME                            <string>    (default: postgres)
  --db-disable-tls/$TEST_DB_DISABLE_TLS              <bool>      (default: false)
  --db-max-idle-conns/$TEST_DB_MAX_IDLE_CONNS        <int>       (default: 2)
  --db-max-open-conns/$TEST_DB_MAX_OPEN_CONNS        <int>       (default: 0)
  --auth-keys-folder/$TEST_AUTH_KEYS_FOLDER          <string>    (default: /service/keys)
  --auth-algorithm/$TEST_AUTH_ALGORITHM              <string>    (default: RS256)
  --zipkin-reporter-uri/$TEST_ZIPKIN_REPORTER_URI    <string>    (default: http://zipkin:9411/api/v2/spans)
  --zipkin-service-name/$TEST_ZIPKIN_SERVICE_NAME    <string>    (default: sales-api)
  --zipkin-probability/$TEST_ZIPKIN_PROBABILITY      <float>     (default: 0.05)
  --help/-h                                          
  display this help message
  --version/-v  
  display version information
`

func TestUsage(t *testing.T) {
	type args struct {
		cfg    interface{}
		prefix string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Cmd", args{cfg: &config.CmdConfig{}, prefix: "TEST"}, cmdConfigHelp, false},
		{"App", args{cfg: &config.AppConfig{}, prefix: "TEST"}, appConfigHelp, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.Usage(tt.args.cfg, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("Usage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Usage() got = %v, want %v", got, tt.want)
				t.Log(cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestParse(t *testing.T) {
	var cmdConfig string = `--version=
--description='testing cmd config'
--args=[migrate]
--db-user='USER'
--db-password=xxxxxx
--db-host='HOST'
--db-name='DB'
--db-disable-tls=false`

	var appConfig string = `--version=
--description='testing app config'
--web-api-host=0.0.0.0:80
--web-debug-host=0.0.0.0:4040
--web-read-timeout=5s
--web-write-timeout=5s
--web-shutdown-timeout=5s
--db-user=admin
--db-password=xxxxxx
--db-host=0.0.0.0:26257
--db-name=postgres
--db-disable-tls=false
--db-max-idle-conns=2
--db-max-open-conns=0
--auth-keys-folder=/service/keys
--auth-algorithm=RS256
--zipkin-reporter-uri=http://zipkin:9411/api/v2/spans
--zipkin-service-name=sales-api
--zipkin-probability=0.01`

	cmdConf := []string{
		"--description='testing cmd config'", "--db-user='USER'",
		"--db-password='V3ryS3cr3t!'", "--db-host='HOST'", "--db-name='DB'",
		"--db-disable-tls=false", "migrate",
	}
	appConf := []string{
		"--description='testing app config'", "--web-api-host=0.0.0.0:80",
		"--web-debug-host=0.0.0.0:4040", "--zipkin-probability=0.01",
		"--db-password='V3ryS3cr3t!'",
	}
	type args struct {
		cfg    interface{}
		prefix string
		args   []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    string
	}{
		{"Empty Cmd", args{cfg: &config.CmdConfig{}, prefix: "TEST", args: []string{}}, false, ""},
		{"Empty cmd -h", args{cfg: &config.CmdConfig{}, prefix: "TEST", args: []string{"-h"}}, true, ""},
		{"Empty cmd --help", args{cfg: &config.CmdConfig{}, prefix: "TEST", args: []string{"--help"}}, true, ""},
		{"Empty cmd -v", args{cfg: &config.CmdConfig{}, prefix: "TEST", args: []string{"-v"}}, true, ""},
		{"Empty cmd --version", args{cfg: &config.CmdConfig{}, prefix: "TEST", args: []string{"--version"}}, true, ""},
		{"Empty App", args{cfg: &config.AppConfig{}, prefix: "TEST", args: []string{}}, false, ""},
		{"Empty App -h", args{cfg: &config.AppConfig{}, prefix: "TEST", args: []string{"-h"}}, true, ""},
		{"Empty App --help", args{cfg: &config.AppConfig{}, prefix: "TEST", args: []string{"--help"}}, true, ""},
		{"Empty App -v", args{cfg: &config.AppConfig{}, prefix: "TEST", args: []string{"-v"}}, true, ""},
		{"Empty App --version", args{cfg: &config.AppConfig{}, prefix: "TEST", args: []string{"--version"}}, true, ""},
		{"Valid Cmd", args{cfg: &config.CmdConfig{}, prefix: "TEST", args: cmdConf}, false, cmdConfig},
		{"Valid App", args{cfg: &config.AppConfig{}, prefix: "TEST", args: appConf}, false, appConfig},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := config.Parse(tt.args.cfg, tt.args.prefix, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, err := conf.String(tt.args.cfg)
			if err != nil {
				t.Error(err)
			}
			if tt.want != "" {
				if got != tt.want {
					t.Errorf("Parse() got = \n%v, want %v", got, tt.want)
					t.Log(cmp.Diff(got, tt.want))
				}
			}
		})
	}
}
