module github.com/tullo/service

go 1.19

// fixes: FAIL - undefined container.StopOptions
// followed the instructions at https://golang.testcontainers.org/quickstart/#2-install-testcontainers-for-go
replace github.com/docker/docker => github.com/docker/docker v20.10.23+incompatible // 22.06 branch

require (
	github.com/alexedwards/argon2id v0.0.0-20211130144151-3585854a6387
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/georgysavva/scany/v2 v2.0.0
	github.com/go-chi/chi/v5 v5.0.8
	github.com/go-playground/locales v0.14.1
	github.com/go-playground/universal-translator v0.18.1
	github.com/go-playground/validator/v10 v10.11.2
	github.com/golang-migrate/migrate/v4 v4.15.2
	github.com/google/go-cmp v0.5.9
	github.com/google/uuid v1.3.0
	github.com/jackc/pgx-zap v0.0.0-20221202020421-94b1cb2f889f
	github.com/jackc/pgx/v5 v5.2.0
	github.com/ory/dockertest/v3 v3.9.1
	github.com/pkg/errors v0.9.1
	github.com/testcontainers/testcontainers-go v0.17.0
	github.com/tullo/conf v1.3.7
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.38.0
	go.opentelemetry.io/otel v1.12.0
	go.opentelemetry.io/otel/exporters/zipkin v1.12.0
	go.opentelemetry.io/otel/sdk v1.12.0
	go.uber.org/zap v1.24.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/Microsoft/hcsshim v0.9.5 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/cockroachdb/cockroach-go/v2 v2.2.20 // indirect
	github.com/containerd/cgroups v1.0.3 // indirect
	github.com/containerd/containerd v1.6.12 // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/docker/cli v20.10.23+incompatible // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.23+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.1.2 // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/patternmatcher v0.5.0 // indirect
	github.com/moby/sys/mount v0.3.3 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2 // indirect
	github.com/opencontainers/runc v1.1.4 // indirect
	github.com/openzipkin/zipkin-go v0.4.1 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/otel/metric v0.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.12.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.5.0 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	golang.org/x/tools v0.1.12 // indirect
	google.golang.org/genproto v0.0.0-20221010155953-15ba04fc1c0e // indirect
	google.golang.org/grpc v1.50.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
