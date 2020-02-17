SHELL := /bin/bash

export PROJECT = ardan-starter-kit
export REGISTRY_ACCOUNT = tullo
export VERSION = 1.0

all: keys sales-api metrics

keys:
	go run ./cmd/sales-admin/main.go keygen private.pem

admin:
	go run ./cmd/sales-admin/main.go --db-disable-tls=1 useradd admin@example.com gophers

migrate:
	go run ./cmd/sales-admin/main.go --db-disable-tls=1 migrate

seed: migrate
	go run ./cmd/sales-admin/main.go --db-disable-tls=1 seed

sales-api:
	docker build \
		-f dockerfile.sales-api \
		-t gcr.io/$(PROJECT)/sales-api-amd64:$(VERSION) \
		--build-arg PACKAGE_NAME=sales-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker image tag \
		gcr.io/$(PROJECT)/sales-api-amd64:$(VERSION) \
		$(REGISTRY_ACCOUNT)/sales-api-amd64:$(VERSION)

metrics:
	docker build \
		-f dockerfile.metrics \
		-t gcr.io/$(PROJECT)/metrics-amd64:$(VERSION) \
		--build-arg PACKAGE_NAME=metrics \
		--build-arg PACKAGE_PREFIX=sidecar/ \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker image tag \
		gcr.io/$(PROJECT)/metrics-amd64:$(VERSION) \
		$(REGISTRY_ACCOUNT)/metrics-amd64:$(VERSION)

up:
	docker-compose up

down:
	docker-compose down

test:
	go test -mod=vendor ./... -count=1

clean:
	docker system prune -f

stop-all:
	docker container stop $$(docker container ls -q --filter "name=sales*" --filter "name=metrics" --filter "name=zipkin")

remove-all:
	docker container rm $$(docker container ls -aq --filter "name=sales*" --filter "name=metrics" --filter "name=zipkin")
deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

deps-upgrade:
	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	go get -t -d -v ./...

deps-cleancache:
	go clean -modcache