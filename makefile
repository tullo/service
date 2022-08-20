SHELL = /bin/bash -o pipefail
# make -n
# make -np 2>&1 | less
export REGISTRY_HOSTNAME = docker.io
export REGISTRY_ACCOUNT = tullo
export VERSION = 1.0
export DOCKER_BUILDKIT = 1
export COMPOSE_DOCKER_CLI_BUILD = 1
export COMPOSE_FILE = deployment/docker/docker-compose.yaml

# ==============================================================================
# Testing the running system: load, traces, metrics)
#
# hey tool: make generate-load
# zipkin: http://localhost:9411
# expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,mem:memstats.Alloc"
#
# To install expvarmon program for metrics dashboard:
# $ go install github.com/divan/expvarmon@latest
#
# To manually generate a private/public key PEM file:
# $ openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# $ openssl rsa -pubout -in private.pem -out public.pem

.DEFAULT_GOAL := run

all: go-run-keygen images run down

images: sales-api metrics

down: compose-down

run: compose-up compose-seed compose-status

staticcheck:
	$$(go env GOPATH)/bin/staticcheck -go 'module' -tests \
		./app/... ./business/... ./foundation/...

staticcheck-install: GO111MODULE := on
staticcheck-install:
	@go install honnef.co/go/tools/cmd/staticcheck@v0.3.3
	@$$(go env GOPATH)/bin/staticcheck -debug.version

go-deps-list:
	go list -mod=mod all

go-deps-reset:
	@git checkout -- go.mod
	@go mod tidy
	@go mod vendor

go-deps-clean-modcache:
	@go clean -modcache

go-deps-upgrade:
#	@go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	@go get -u -t -d -v ./...
	@go mod tidy
	@go mod vendor

go-tidy:
	go mod tidy
	go mod vendor

go-run-api:
	@go run ./app/sales-api \
		--db-disable-tls=1 \
		--auth-keys-folder=deployment/keys \
		--zipkin-reporter-uri=http://0.0.0.0:9411/api/v2/spans \
		--zipkin-probability=1

go-run-config:
	@go run ./app/sales-api -h

go-run-keygen:
	@go run ./app/sales-admin/main.go keygen

go-run-tokengen: USERID=5cf37266-3473-4006-984f-9325122678b7
go-run-tokengen: SIGNING_KEY_ID=deployment/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem
go-run-tokengen: ALGORITHM=RS256
go-run-tokengen: go-run-migrate
	@echo tokengen \(userID, privateKeyPEM, algorithm\)
	@go run ./app/sales-admin/main.go --db-disable-tls=1 tokengen ${USERID} ${SIGNING_KEY_ID} ${ALGORITHM}

go-run-migrate: CMD=/cockroach/cockroach node status --insecure
go-run-migrate: compose-db-up
	@docker-compose -f $(COMPOSE_FILE) exec db sh -c 'until ${CMD}; do { printf '.'; sleep 1; }; done'
	@go run ./app/sales-admin/main.go --db-disable-tls=1 migrate

go-run-seed: go-run-migrate
	@go run ./app/sales-admin/main.go --db-disable-tls=1 seed

go-run-useradd: go-run-migrate
	@go run ./app/sales-admin/main.go --db-disable-tls=1 useradd admin admin@example.com gophers

go-run-users: go-run-migrate
	@go run ./app/sales-admin/main.go --db-disable-tls=1 users 1 50

go-pprof-browser:
	@firefox http://localhost:4000/debug/pprof

go-pprof-heap:
	@go tool pprof http://localhost:4000/debug/pprof/heap

go-pprof-profile:
	@go tool pprof http://localhost:4000/debug/pprof/profile?seconds=30
#   (pprof) top10 -cum

go-test: staticcheck
	@go vet ./app/... ./business/... ./foundation/...
	@go test ./... -count=1
#	@go test -v ./... -count=1
#	@go test -v -run TestProducts ./app/sales-api/tests/ -count=1
#	@go test -v -run TestProducts/crudProductUser ./app/sales-api/tests/ -count=1

compose-config:
	@docker-compose -f $(COMPOSE_FILE) config

compose-down:
	@docker-compose -f $(COMPOSE_FILE) down --remove-orphans --volumes

compose-logs:
	@docker-compose -f $(COMPOSE_FILE) logs -f --tail="30"

compose-migrate:
	@docker-compose -f $(COMPOSE_FILE) exec sales-api /service/admin migrate

compose-seed: compose-migrate
	@docker-compose -f $(COMPOSE_FILE) exec sales-api /service/admin seed

compose-status:
	@docker-compose -f $(COMPOSE_FILE) ps --all

compose-tokengen: USERID=5cf37266-3473-4006-984f-9325122678b7
compose-tokengen: PRIVATE_KEY_FILE=/service/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem
compose-tokengen: ALGORITHM=RS256
compose-tokengen: compose-migrate
	@docker-compose -f $(COMPOSE_FILE) exec sales-api /service/admin tokengen ${USERID} "${PRIVATE_KEY_FILE}" ${ALGORITHM}

compose-up: CMD=/cockroach/cockroach node status --insecure
compose-up:
	@docker-compose -f $(COMPOSE_FILE) up --detach --remove-orphans
	@docker-compose -f $(COMPOSE_FILE) exec db sh -c 'until ${CMD}; do { printf '.'; sleep 1; }; done'

compose-db-up:
	@docker-compose -f $(COMPOSE_FILE)  up --detach --remove-orphans db

compose-db-shell: USER=postgres
compose-db-shell: compose-db-up
	@docker-compose -f $(COMPOSE_FILE) exec db psql --username=${USER} --dbname=${USER}

curl-readiness-check:
	@curl -i --silent --show-error http://0.0.0.0:4000/debug/readiness
	@echo

curl-liveness-check:
	@curl -i --silent --show-error http://0.0.0.0:4000/debug/liveness
	@echo

curl-jwt-token:
	SIGNING_KEY_ID=54bb2165-71e1-41a6-af3e-7da4a0e1e2c1; \
	curl --no-progress-meter --user "admin@example.com:gophers" http://localhost:3000/v1/users/token/$${SIGNING_KEY_ID} | jq

curl-users:
	SIGNING_KEY_ID=54bb2165-71e1-41a6-af3e-7da4a0e1e2c1; \
	TOKEN=$$(curl --no-progress-meter --user 'admin@example.com:gophers' http://localhost:3000/v1/users/token/$${SIGNING_KEY_ID} | jq -r '.token'); \
	curl --no-progress-meter -H "Authorization: Bearer $${TOKEN}" http://0.0.0.0:3000/v1/users/1/50 | jq

curl-products:
	SIGNING_KEY_ID=54bb2165-71e1-41a6-af3e-7da4a0e1e2c1; \
	TOKEN=$$(curl --no-progress-meter --user 'admin@example.com:gophers' http://localhost:3000/v1/users/token/$${SIGNING_KEY_ID} | jq -r '.token'); \
	curl --no-progress-meter -H "Authorization: Bearer $${TOKEN}" http://0.0.0.0:3000/v1/products/1/50 | jq

.PHONY: generate-load
generate-load: export SIGNING_KEY_ID=54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
generate-load: export TOKEN=$$(curl --no-progress-meter --user 'admin@example.com:gophers' \
	http://localhost:3000/v1/users/token/${SIGNING_KEY_ID} | jq -r '.token')
generate-load:
	@wget -q -O - --header "Authorization: Bearer $(TOKEN)" http://localhost:3000/v1/products/1/50 | jq
	@echo "Running 'hey' tool: sending 100'000 requests via 50 concurrent workers."
	@$$(go env GOPATH)/bin/hey -c 50 -n 100000 -H "Authorization: Bearer $(TOKEN)" http://localhost:3000/v1/products/1/50

# https://www.gnu.org/software/make/manual/make.html#Target_002dspecific

hey-upgrade: GO111MODULE := on
hey-upgrade:
	@go install -v github.com/rakyll/hey@latest
	$$(go env GOPATH)/bin/hey

metrics:
	@docker buildx build \
		-f deployment/docker/dockerfile.metrics \
		-t $(REGISTRY_HOSTNAME)/$(REGISTRY_ACCOUNT)/metrics-amd64:$(VERSION) \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

sales-api:
	@docker buildx build \
		-f deployment/docker/dockerfile.sales-api \
		-t $(REGISTRY_HOSTNAME)/$(REGISTRY_ACCOUNT)/sales-api-amd64:$(VERSION) \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

docker-stop-all:
	@docker container stop $$(docker container ls -q --filter "name=sales*" --filter "name=metrics" --filter "name=zipkin")

docker-remove-all:
	@docker container rm $$(docker container ls -aq --filter "name=sales*" --filter "name=metrics" --filter "name=zipkin")

docker-prune-system:
	@docker system prune -f
	
docker-prune-build-cache:
	@docker buildx prune -f

docker-buildx-install:
	./buildx-install.sh
