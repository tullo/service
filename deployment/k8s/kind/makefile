SHELL = /bin/bash -o pipefail

export PROJECT = tullo-starter-kit
export CLUSTER = tullo-starter-cluster
export VERSION = 1.0

all:
	(cd ../.. && make all)

sales-api:
	(cd ../.. && make sales-api)

metrics:
	(cd ../.. && make metrics)

test:
	(cd ../.. && make test)

up:
	kind create cluster --name $(CLUSTER) --config kind-config.yaml

down:
	kind delete cluster --name $(CLUSTER)

load:
	kind load docker-image gcr.io/$(PROJECT)/sales-api-amd64:$(VERSION) --name $(CLUSTER)
	kind load docker-image gcr.io/$(PROJECT)/metrics-amd64:$(VERSION) --name $(CLUSTER)
	# kind load docker-image openzipkin/zipkin-slim:2.21 --name $(CLUSTER)
	# kind load docker-image postgres:12.4-alpine --name $(CLUSTER)

services:
	kubectl create -f deploy-postgres.yaml
	kubectl create -f deploy-sales-api.yaml
	@echo ======================================================================

update-sales-api:
	kind load docker-image gcr.io/$(PROJECT)/sales-api-amd64:$(VERSION) --name $(CLUSTER)
	kubectl set image pod <POD_NAME> sales-api=gcr.io/$(PROJECT)/sales-api-amd64:$(VERSION)
	kubectl delete pod <POD_NAME>

get-pods:
	kubectl get pods

logs:
	kubectl logs <POD_NAME> sales-api

status:
	kubectl get nodes
	kubectl get pods
	kubectl get services sales-api
	kubectl logs <POD_ID> -c sales-api
	@echo ======================================================================

shell:
	# kubectl get pods
	# kubectl exec -it <POD NAME> --container sales-api  -- /bin/sh
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed
	# curl --user "admin@example.com:gophers" http://localhost:3000/v1/users/token
	# export TOKEN="COPY TOKEN STRING FROM LAST CALL"
	# curl -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users
	@echo ======================================================================

delete:
	kubectl delete services sales-api
	kubectl delete deployment sales-api	
	@echo ======================================================================