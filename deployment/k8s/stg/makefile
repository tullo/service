SHELL = /bin/bash -o pipefail

export PROJECT = tullo-starter-kit
export CLUSTER = tullo-starter-cluster
export CONTAINER_REGISTRY = eu.gcr.io
export DATABASE = tullo-starter-db
export ZONE = europe-west3
export VERSION = 1.0

.DEFAULT_GOAL := contexts

contexts:
	@kubectl config get-contexts

use-context:
	@kubectl config use-context gke_$(PROJECT)_europe-west3_$(CLUSTER)

cluster-create:
	$(shell go env GOPATH)/bin/kind create cluster --name $(CLUSTER) --config kind-config.yaml

cluster-delete:
	$(shell go env GOPATH)/bin/kind delete cluster --name $(CLUSTER)

cluster-info:
	@kubectl cluster-info --context gke_$(PROJECT)_europe-west3_$(CLUSTER)

# https://cloud.google.com/container-registry/docs/pushing-and-pulling
image-registry:
	gcloud auth configure-docker $(CONTAINER_REGISTRY)

docker-tag-images:
	@docker tag tullo/sales-api-amd64:$(VERSION) $(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64:$(VERSION)
	@docker tag tullo/metrics-amd64:$(VERSION) $(CONTAINER_REGISTRY)/$(PROJECT)/metrics-amd64:$(VERSION)

docker-push-images:
	@docker image push $(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64:$(VERSION)
	@docker image push $(CONTAINER_REGISTRY)/$(PROJECT)/metrics-amd64:$(VERSION)

# https://cloud.google.com/sdk/gcloud/reference/container/images/list
image-list:
	@gcloud container images list --repository $(CONTAINER_REGISTRY)/$(PROJECT)

# https://cloud.google.com/sdk/gcloud/reference/container/images/list-tags
image-tag-list:
	@echo '==>' listing tags for image: [$(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64]:
	@gcloud container images list-tags $(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64
	@echo
	@echo '==>' listing tags for image: [$(CONTAINER_REGISTRY)/$(PROJECT)/metrics-amd64]:
	@gcloud container images list-tags $(CONTAINER_REGISTRY)/$(PROJECT)/metrics-amd64

kubeval:
	@kustomize build . | kubeval --strict --force-color -

deployment-apply: kubeval
	@kustomize build . | kubectl apply --validate -f -
#	@kustomize build . | kubectl apply --dry-run=client --validate -f -
	@watch kubectl get po

deployment-delete:
	kustomize build . | kubectl delete -f -

update-sales-api:
	@kubectl set image deployment -l app=sales-api sales-api=$(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64:$(VERSION)

get-pods:
	@kubectl get pods

desc-app-pod:
	@$(eval APP=`kubectl get pod -l app=sales-api -o jsonpath='{.items[0].metadata.name}'`)
	@kubectl describe po/${APP}

.PHONY: logs
logs:
	@echo ====== metrics ==========================================================
	@kubectl logs --tail=8 -l app=sales-api --container metrics
	@echo
	@echo ====== sales-api ========================================================
	@kubectl logs --tail=1 -l app=sales-api --container app
	@echo
	@echo ====== zipkin ===========================================================
	@kubectl logs --tail=10 -l app=sales-api --container zipkin

schema: get-pods migrate seed

migrate:
	@$(eval APP=`kubectl get pod -l app=sales-api -o jsonpath='{.items[0].metadata.name}'`)
	@kubectl exec -it ${APP} --container app  -- /app/admin --db-disable-tls=1 migrate

seed:
	@$(eval APP=`kubectl get pod -l app=sales-api -o jsonpath='{.items[0].metadata.name}'`)
	@kubectl exec -it ${APP} --container app  -- /app/admin --db-disable-tls=1 seed

health-request:
	@echo ====== sales-api =====================================================
	@wget -q -O - http://sales-api:3000/v1/health | jq

users-request:
	@$(eval TOKEN=`curl --no-progress-meter --user 'admin@example.com:gophers' \
		http://localhost:3000/v1/users/token | jq -r '.token'`)
	@wget -q -O - --header "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users | jq

status:
	@echo ====== nodes =========================================================
	@kubectl get nodes
	@echo ====== pods ==========================================================
	@kubectl get pods -o wide
	@echo ====== services ======================================================
	@kubectl get services sales-api -o wide
	@echo ====== deploy ========================================================
	@kubectl get deploy/sales-api -o wide
	@echo ====== replicaset ====================================================
	@kubectl get rs -l app=sales-api -o wide
	@echo ====== statefulset ===================================================
	@kubectl get sts -l app=sales-api -o wide

sales-api-shell: get-pods
	@$(eval APP=`kubectl get pod -l app=sales-api -o jsonpath='{.items[0].metadata.name}'`)
#	@kubectl exec -it ${APP} --container app  -- /bin/sh
	@kubectl exec -it ${APP} --container app  -- env | grep DB
	@kubectl exec -it ${APP} --container app  -- wget -q -O - http://sales-api:3000/v1/health

port-forward: get-pods
	@$(eval APP=`kubectl get pod -l app=sales-api -o jsonpath='{.items[0].metadata.name}'`)
	@kubectl port-forward pod/${APP} :3000

database:
	@gcloud sql instances create $(DATABASE) --database-version=POSTGRES_12 --tier=db-f1-micro --zone=$(ZONE)-a
	@gcloud sql instances describe $(DATABASE)
