module github.com/tullo/service

go 1.16

require (
	github.com/alexedwards/argon2id v0.0.0-20210326052512-e2135f7c9c77
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/go-chi/chi v1.5.4
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator/v10 v10.6.0
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.2.0
	github.com/jmoiron/sqlx v1.3.3
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lib/pq v1.10.1
	github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/tullo/conf v1.3.7
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.16.0
	go.opentelemetry.io/otel v0.16.0
	go.opentelemetry.io/otel/exporters/trace/zipkin v0.16.0
	go.opentelemetry.io/otel/sdk v0.16.0
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf // indirect
	golang.org/x/sys v0.0.0-20210507014357-30e306a8bba5 // indirect
	golang.org/x/text v0.3.6 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
)

// replace github.com/golang-migrate/migrate/v4 => ../migrate/
