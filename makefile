SHELL := /bin/bash

export PROJECT = ardan-starter-kit
export REGISTRY_HOSTNAME = docker.io
export REGISTRY_ACCOUNT = tullo
export VERSION = 1.0
export DOCKER_BUILDKIT = 1

all: go-run-keygen sales-api metrics run down

down: compose-down

run: compose-up compose-seed compose-status

staticcheck:
	$(shell go env GOPATH)/bin/staticcheck -go 1.14 -tests ./cmd/... ./internal/...

staticcheck-upgrade:
	GO111MODULE=off go get -u honnef.co/go/tools/cmd/staticcheck
	$(shell go env GOPATH)/bin/staticcheck -debug.version

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

go-deps-clean-modcache:
	go clean -modcache

go-deps-upgrade:
#	go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	go get -u -t -d -v ./...

go-tidy:
	go mod tidy
	go mod vendor

go-run-admin-user: go-run-migrate
	go run ./cmd/sales-admin/main.go --db-disable-tls=1 useradd admin@example.com gophers

go-run-api:
	go run ./cmd/sales-api --db-disable-tls=1 --auth-private-key-file=private.pem

go-run-keygen:
	go run ./cmd/sales-admin/main.go keygen private.pem

go-run-migrate: compose-db-up
	docker-compose exec db sh -c 'until $$(nc -z localhost 5432); do { printf '.'; sleep 1; }; done'
	go run ./cmd/sales-admin/main.go --db-disable-tls=1 migrate

go-run-seed: go-run-migrate
	go run ./cmd/sales-admin/main.go --db-disable-tls=1 seed

go-pprof-heap:
	go tool pprof http://localhost:4000/debug/pprof/heap
#	firefox http://localhost:4000/debug/pprof
#	go tool pprof http://localhost:4000/debug/pprof/profile?seconds=30
	@echo

go-test: staticcheck
	go vet ./cmd/... ./internal/...
	go test ./... -count=1
#	go test -v ./... -count=1
#	go test -v -run TestProducts ./cmd/sales-api/tests/ -count=1
#	go test -v -run TestProducts/crudProductUser ./cmd/sales-api/tests/ -count=1

compose-db-up:
	docker-compose up --detach --remove-orphans db

compose-down:
	docker-compose down --remove-orphans --volumes

compose-logs:
	docker-compose logs -f

compose-migrate:
	docker-compose exec sales-api /app/admin migrate

compose-seed: compose-migrate
	docker-compose exec sales-api /app/admin seed

compose-status:
	docker-compose ps --all

compose-up:
	docker-compose up --detach --remove-orphans

curl-health-check:
	curl -v http://0.0.0.0:3000/v1/health | jq
	@echo

generate-load:
	$(TOKEN = $(shell curl --user "admin@example.com:gophers" http://localhost:3000/v1/users/token | jq -r '.token'))
#	@echo $(TOKEN)
#	curl -s -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/products | jq
	hey -c 10 -n 30000 -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/products
	@echo

metrics:
	docker build \
		-f dockerfile.metrics \
		-t $(REGISTRY_HOSTNAME)/$(REGISTRY_ACCOUNT)/metrics-amd64:$(VERSION) \
		--build-arg PACKAGE_NAME=metrics \
		--build-arg PACKAGE_PREFIX=sidecar/ \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker image tag \
		$(REGISTRY_ACCOUNT)/metrics-amd64:$(VERSION) \
		gcr.io/$(PROJECT)/metrics-amd64:$(VERSION)

sales-api:
	docker build \
		-f dockerfile.sales-api \
		-t $(REGISTRY_HOSTNAME)/$(REGISTRY_ACCOUNT)/sales-api-amd64:$(VERSION) \
		--build-arg PACKAGE_NAME=sales-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker image tag \
		$(REGISTRY_ACCOUNT)/sales-api-amd64:$(VERSION) \
		gcr.io/$(PROJECT)/sales-api-amd64:$(VERSION)

docker-stop-all:
	docker container stop $$(docker container ls -q --filter "name=sales*" --filter "name=metrics" --filter "name=zipkin")

docker-remove-all:
	docker container rm $$(docker container ls -aq --filter "name=sales*" --filter "name=metrics" --filter "name=zipkin")

docker-prune-system:
	docker system prune -f

#==============================================================================
# GKE

gcloud-config:
	@echo Setting environment for ardan-starter-kit
	gcloud config set project ardan-starter-kit
	gcloud config set compute/zone us-central1-b
	gcloud auth configure-docker
	@echo ======================================================================

gcloud-project:
	gcloud projects create ardan-starter-kit
	gcloud beta billing projects link ardan-starter-kit --billing-account=$(ACCOUNT_ID)
	gcloud services enable container.googleapis.com
	@echo ======================================================================

gcloud-cluster:
	gcloud container clusters create ardan-starter-cluster --enable-ip-alias --num-nodes=2 --machine-type=n1-standard-2
	gcloud compute instances list
	@echo ======================================================================

gcloud-upload:
	docker push gcr.io/ardan-starter-kit/sales-api-amd64:1.0
	docker push gcr.io/ardan-starter-kit/metrics-amd64:1.0
	@echo ======================================================================

gcloud-network:
	# Creating your own VPC network. We will use the default VPC.
	gcloud compute networks create ardan-starter-vpc --subnet-mode=auto --bgp-routing-mode=regional
	gcloud compute addresses create ardan-starter-network --global --purpose=VPC_PEERING --prefix-length=16 --network=ardan-starter-vpc
	gcloud compute addresses list --global --filter="purpose=VPC_PEERING"
	@echo ======================================================================

gcloud-database:
	gcloud beta sql instances create ardan-starter-db --database-version=POSTGRES_9_6 --no-backup --tier=db-f1-micro --zone=us-central1-b --no-assign-ip --network=default
	gcloud sql instances describe ardan-starter-db
	@echo ======================================================================

gcloud-db-assign-ip:
	gcloud sql instances patch ardan-starter-db --authorized-networks=[YOUR_IP/32]
	gcloud sql instances describe ardan-starter-db
	@echo ======================================================================

gcloud-services:
	# Make sure the deploy script has the right IP address for the DB.
	kubectl create -f gke-deploy-sales-api.yaml
	kubectl expose -f gke-expose-sales-api.yaml --type=LoadBalancer
	@echo ======================================================================

gcloud-status:
	gcloud container clusters list
	kubectl get nodes
	kubectl get pods
	kubectl get services sales-api
	@echo ======================================================================

gcloud-shell:
	# kubectl get pods
	kubectl exec -it <POD NAME> --container sales-api  -- /bin/sh
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed
	@echo ======================================================================

gcloud-delete:
	kubectl delete services sales-api
	kubectl delete deployment sales-api	
	gcloud container clusters delete sales-api-cluster
	gcloud projects delete sales-api
	gcloud container images delete gcr.io/ardan-starter-kit/sales-api-amd64:1.0 --force-delete-tags
	gcloud container images delete gcr.io/ardan-starter-kit/metrics-amd64:1.0 --force-delete-tags
	docker image remove gcr.io/sales-api/sales-api-amd64:1.0
	docker image remove gcr.io/sales-api/metrics-amd64:1.0
	@echo ======================================================================

#===============================================================================
# GKE Installation
#
# Install the Google Cloud SDK. This contains the gcloud client needed to perform
# some operatings
# https://cloud.google.com/sdk/
#
# Installing the K8s kubectl client. 
# https://kubernetes.io/docs/tasks/tools/install-kubectl/

# ==============================================================================
# make debuging
# make -n
# make -np 2>&1 | less
