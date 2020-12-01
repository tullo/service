SHELL = /bin/bash -o pipefail
# make -n
# make -np 2>&1 | less
export PROJECT = tullo-starter-kit
export REGISTRY_HOSTNAME = docker.io
export REGISTRY_ACCOUNT = tullo
export VERSION = 1.0
export DOCKER_BUILDKIT = 1
export COMPOSE_DOCKER_CLI_BUILD = 1
export COMPOSE_FILE = deployment/docker/docker-compose.yaml

# ==============================================================================
# Testing running system
#
# hey tool: make generate-load
# zipkin: http://localhost:9411
# expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,mem:memstats.Alloc"
#
# To manually generate a private/public key PEM file:
# 1. openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# 2. openssl rsa -pubout -in private.pem -out public.pem

# https://www.gnu.org/software/make/manual/make.html#Target_002dspecific
hey-upgrade: export GO111MODULE := off

.DEFAULT_GOAL := run

all: go-run-keygen images run down

images: sales-api metrics

down: compose-down

run: compose-up compose-seed compose-status

check:
	$(shell go env GOPATH)/bin/staticcheck -go 1.15 \
		-tests ./app/... ./business/... ./foundation/...

clone:
	@git clone git@github.com:dominikh/go-tools.git /tmp/go-tools \
		&& cd /tmp/go-tools \
		&& git checkout "2020.1.6" \

install:
	@cd /tmp/go-tools && go install -v ./cmd/staticcheck
	$(shell go env GOPATH)/bin/staticcheck -debug.version

.PHONY: staticcheck
staticcheck: clone install

deps-reset:
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
		--auth-private-key-file=private.pem \
		--zipkin-reporter-uri=http://0.0.0.0:9411/api/v2/spans \
		--zipkin-probability=1

go-run-config:
	@go run ./app/sales-api -h

go-run-keygen:
	@go run ./app/sales-admin/main.go keygen

go-run-tokengen: go-run-migrate
	@echo tokengen \(userID, privateKeyPEM, algorithm\)
	@go run ./app/sales-admin/main.go --db-disable-tls=1 tokengen '5cf37266-3473-4006-984f-9325122678b7' private.pem 'RS256'

go-run-migrate: compose-db-up
	@docker-compose -f $(COMPOSE_FILE) exec db sh -c 'until $$(nc -z localhost 5432); do { printf '.'; sleep 1; }; done'
	@go run ./app/sales-admin/main.go --db-disable-tls=1 migrate

go-run-seed: go-run-migrate
	@go run ./app/sales-admin/main.go --db-disable-tls=1 seed

go-run-useradd: go-run-migrate
	@go run ./app/sales-admin/main.go --db-disable-tls=1 useradd admin@example.com gophers

go-run-users: go-run-migrate
	@go run ./app/sales-admin/main.go --db-disable-tls=1 users 1 50

go-pprof-browser:
	@firefox http://localhost:4000/debug/pprof

go-pprof-heap:
	@go tool pprof http://localhost:4000/debug/pprof/heap

go-pprof-profile:
	@go tool pprof http://localhost:4000/debug/pprof/profile?seconds=30
#   (pprof) top10 -cum

go-test: check
	@go vet ./app/... ./business/... ./foundation/...
	@go test ./... -count=1
#	@go test -v ./... -count=1
#	@go test -v -run TestProducts ./app/sales-api/tests/ -count=1
#	@go test -v -run TestProducts/crudProductUser ./app/sales-api/tests/ -count=1

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

compose-up:
	@docker-compose -f $(COMPOSE_FILE) up --detach --remove-orphans
	@docker-compose -f $(COMPOSE_FILE) exec db sh -c 'until $$(nc -z localhost 5432); do { printf '.'; sleep 1; }; done'

compose-db-up:
	@docker-compose -f $(COMPOSE_FILE)  up --detach --remove-orphans db

curl-readiness-check:
	@curl -i --silent --show-error http://0.0.0.0:3000/v1/readiness
	@echo

curl-liveness-check:
	@curl -i --silent --show-error http://0.0.0.0:3000/v1/liveness
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
generate-load:
	@$(eval TOKEN=`curl --no-progress-meter --user 'admin@example.com:gophers' \
		http://localhost:3000/v1/users/token | jq -r '.token'`)
	@wget -q -O - --header "Authorization: Bearer $(TOKEN)" http://localhost:3000/v1/products | jq
	@echo "Running 'hey' tool: sending 10'000 requests via 100 concurrent workers."
	@$(shell go env GOPATH)/bin/hey -c 10 -n 10000 -H "Authorization: Bearer $(TOKEN)" http://localhost:3000/v1/products

.PHONY: hey-upgrade
hey-upgrade:
	@echo GO111MODULE=$(GO111MODULE)
	@go get -u -v github.com/rakyll/hey
	$(shell go env GOPATH)/bin/hey

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
