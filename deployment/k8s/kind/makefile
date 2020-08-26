SHELL = /bin/bash -o pipefail

export PROJECT = tullo-starter-kit
export CLUSTER = tullo-starter-cluster
export VERSION = 1.0

cluster-create:
	$(shell go env GOPATH)/bin/kind create cluster --name $(CLUSTER) --config kind-config.yaml

cluster-delete:
	$(shell go env GOPATH)/bin/kind delete cluster --name $(CLUSTER)

cluster-info:
	@kubectl cluster-info --context kind-$(CLUSTER)

images-load:
	@$(shell go env GOPATH)/bin/kind load docker-image eu.gcr.io/$(PROJECT)/sales-api-amd64:$(VERSION) --name $(CLUSTER)
	@$(shell go env GOPATH)/bin/kind load docker-image eu.gcr.io/$(PROJECT)/metrics-amd64:$(VERSION) --name $(CLUSTER)

images-list:
	@docker exec -it $(CLUSTER)-control-plane crictl images

deployment-create:
	@kubectl create -f deploy-postgres.yaml
	@kubectl create -f deploy-sales-api.yaml

deployment-delete:
	@kubectl delete -f deploy-sales-api.yaml
	@kubectl delete -f deploy-postgres.yaml

update-sales-api:
	@$(shell go env GOPATH)/bin/kind load docker-image eu.gcr.io/$(PROJECT)/sales-api-amd64:$(VERSION) --name $(CLUSTER)
	@kubectl set image pod <POD_NAME> sales-api=eu.gcr.io/$(PROJECT)/sales-api-amd64:$(VERSION)

get-pods:
	@kubectl get pods

.PHONY: logs
logs:
	@echo ====== postgres =========================================================
	@kubectl logs --tail=5 <POD_NAME>
	@echo
	@echo ====== metrics ==========================================================
	@kubectl logs --tail=8 <POD_NAME> --container metrics
	@echo
	@echo ====== sales-api ========================================================
	@kubectl logs --tail=1 <POD_NAME> --container sales-api
	@echo
	@echo ====== zipkin ===========================================================
	@kubectl logs --tail=10 <POD_NAME> --container zipkin

schema: get-pods migrate seed

migrate:
	@kubectl exec -it <POD_NAME> --container sales-api  -- /app/admin --db-disable-tls=1 migrate

seed:
	@kubectl exec -it <POD_NAME> --container sales-api  -- /app/admin --db-disable-tls=1 seed

health-request:
	@echo ====== postgres ======================================================
	@kubectl exec -it <POD_NAME> -- pg_isready
	@echo 
	@echo ====== sales-api =====================================================
	@wget -q -O - http://localhost:3000/v1/health | jq

users-request:
	@$(eval TOKEN=`curl --no-progress-meter --user 'admin@example.com:gophers' \
		http://localhost:3000/v1/users/token | jq -r '.token'`)
	@wget -q -O - --header "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users | jq

status:
	@echo ====== nodes =========================================================
	@kubectl get nodes
	@echo ====== pods ==========================================================
	@kubectl get pods
	@echo ====== services ======================================================
	@kubectl get services postgres sales-api

sales-api-shell: get-pods
	@kubectl exec -it <POD_NAME> --container sales-api  -- /bin/sh
