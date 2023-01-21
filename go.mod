module github.com/tullo/service

go 1.19

require (
	github.com/alexedwards/argon2id v0.0.0-20211130144151-3585854a6387
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/georgysavva/scany/v2 v2.0.0
	github.com/go-chi/chi/v4 v4.1.2+incompatible
	github.com/go-playground/locales v0.14.1
	github.com/go-playground/universal-translator v0.18.0
	github.com/go-playground/validator/v10 v10.11.1
	github.com/golang-migrate/migrate/v4 v4.15.1
	github.com/google/go-cmp v0.5.9
	github.com/google/uuid v1.3.0
	github.com/jackc/pgconn v1.13.0
	github.com/jackc/pgx/v5 v5.0.0
	github.com/ory/dockertest/v3 v3.9.1
	github.com/pkg/errors v0.9.1
	github.com/tullo/conf v1.3.7
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.16.0
	go.opentelemetry.io/otel v0.16.0
	go.opentelemetry.io/otel/exporters/trace/zipkin v0.16.0
	go.opentelemetry.io/otel/sdk v0.16.0
	gopkg.in/go-playground/validator.v9 v9.31.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/cockroachdb/cockroach-go/v2 v2.2.19 // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/docker/cli v20.10.22+incompatible // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.22+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/puddle/v2 v2.0.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lib/pq v1.10.6 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.1.2 // indirect
	github.com/openzipkin/zipkin-go v0.2.5 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.opentelemetry.io/contrib v0.20.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.4.0 // indirect
	golang.org/x/net v0.4.0 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// TODO: switch to pgx once it gets released
// https://github.com/golang-migrate/migrate/tree/master/database/pgx
// go mod why github.com/lib/pq
// github.com/tullo/service/business/data/schema
// github.com/golang-migrate/migrate/v4/database/postgres
// github.com/lib/pq

// replace github.com/golang-migrate/migrate/v4 => ../migrate/
