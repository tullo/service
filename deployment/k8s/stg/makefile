#!make
SHELL = /bin/bash -o pipefail
VERSION = 1.0
PROJECT = stackwise-starter-kit
CLUSTER = stackwise-starter-cluster
CONTAINER_REGISTRY = eu.gcr.io
REGION = europe-west3
ZONE = europe-west3-c
DATABASE_INSTANCE = stackwise-starter-db-$(REGION)
DB_ROOT_PASSWORD = r00tpasswd
DATABASE = stackwise
PRIVATE_ADDRESS = 'make gc-db-instance-list & put PRIVATE_ADDRESS here'
MY_IP = 12.34.56.78
include .env
#export

.DEFAULT_GOAL := kctl-config-contexts

kctl-config-contexts:
	@kubectl config get-contexts

kctl-config-use-context:
	@kubectl config use-context gke_$(PROJECT)_$(ZONE)_$(CLUSTER)

kctl-cluster-info:
	@kubectl cluster-info --context gke_$(PROJECT)_$(ZONE)_$(CLUSTER)

# https://cloud.google.com/container-registry/docs/pushing-and-pulling
# Register gcloud as the credential helper for the Google-supported Docker registry.
gc-auth-configure-docker:
	gcloud auth configure-docker $(CONTAINER_REGISTRY)

docker-tag-images:
	@docker tag tullo/sales-api-amd64:$(VERSION) $(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64:$(VERSION)
	@docker tag tullo/metrics-amd64:$(VERSION) $(CONTAINER_REGISTRY)/$(PROJECT)/metrics-amd64:$(VERSION)

docker-push-images:
	@docker image push $(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64:$(VERSION)
	@docker image push $(CONTAINER_REGISTRY)/$(PROJECT)/metrics-amd64:$(VERSION)

# https://cloud.google.com/sdk/gcloud/reference/container/images/list
gc-images-list:
	@gcloud container images list --repository $(CONTAINER_REGISTRY)/$(PROJECT)

# https://cloud.google.com/sdk/gcloud/reference/container/images/list-tags
gc-images-list-tags:
	@echo '==>' listing tags for image: [$(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64]:
	@gcloud container images list-tags $(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64
	@echo
	@echo '==>' listing tags for image: [$(CONTAINER_REGISTRY)/$(PROJECT)/metrics-amd64]:
	@gcloud container images list-tags $(CONTAINER_REGISTRY)/$(PROJECT)/metrics-amd64

# validate kubernetes yaml manifests
kubeval:
	@kustomize build . | kubeval --strict --force-color -

deployment-apply: kubeval
	@kustomize build . | kubectl apply --validate -f -
#	@kustomize build . | kubectl apply --dry-run=client --validate -f -
	@watch kubectl get po

deployment-delete:
	kustomize build . | kubectl delete -f -

kctl-update-sales-api-image:
	@kubectl set image deployment -l app=sales-api sales-api=$(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64:$(VERSION)

kctl-get-pods:
	@kubectl get pods

kctl-describe-app-pod:
	@$(eval APP=`kubectl get pod -l app=sales-api -o jsonpath='{.items[0].metadata.name}'`)
	@kubectl describe po/${APP}

.PHONY: kctl-logs
kctl-logs:
	@echo ====== metrics log ==================================================
	@kubectl logs --tail=8 -l app=sales-api --container metrics
	@echo
	@echo ====== sales-api log ================================================
	@kubectl logs --tail=1 -l app=sales-api --container app
	@echo
	@echo ====== zipkin log ===================================================
	@kubectl logs --tail=10 -l app=sales-api --container zipkin

app-shell: kctl-get-pods
	@$(eval APP=`kubectl get pod -l app=sales-api -o jsonpath='{.items[0].metadata.name}'`)
	@kubectl exec -it ${APP} --container app  -- sh
#	@kubectl exec -it ${APP} --container app  -- env | grep DB

app-schema: kctl-get-pods app-migrate app-seed

app-migrate:
	@$(eval APP=`kubectl get pod -l app=sales-api -o jsonpath='{.items[0].metadata.name}'`)
	@kubectl exec -it ${APP} --container app  -- /app/admin --db-disable-tls=1 migrate

app-seed:
	@$(eval APP=`kubectl get pod -l app=sales-api -o jsonpath='{.items[0].metadata.name}'`)
	@kubectl exec -it ${APP} --container app  -- /app/admin --db-disable-tls=1 seed

app-health-request:
	@echo ====== /v1/health ===================================================
	@$(eval APP=`kubectl get pod -l app=sales-api -o jsonpath='{.items[0].metadata.name}'`)
	@kubectl exec -it ${APP} --container app  -- wget -q -O - http://localhost:3000/v1/health

lb-health-request:
	@echo ====== health-request via LoadBalancer ==============================
	@$(eval IP=`kubectl get svc sales-api -o wide -o jsonpath='{.status.loadBalancer.ingress[0].ip}'`)
	@echo IP: ${IP}
	@wget -q -O - http://${IP}:3000/v1/health | jq

# make -np 2>&1 users-request | less
lb-users-request:
	@echo ====== users-request via LoadBalancer ===============================
	@$(eval IP=`kubectl get svc sales-api -o wide -o jsonpath='{.status.loadBalancer.ingress[0].ip}'`)
	@$(eval TOKEN=$(shell curl --no-progress-meter --user 'admin@example.com:gophers' http://${IP}:3000/v1/users/token | jq -r '.token'))
	@wget -q -O - --header "Authorization: Bearer ${TOKEN}" http://${IP}:3000/v1/users | jq


kctl-status:
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

# Listen on a random port locally, forwarding to 3000 in the pod
# - kubectl port-forward pod/${APP} :3000
# kubectl port-forward -h
kctl-port-forward: kctl-get-pods
	kubectl port-forward service/sales-api 3000

kctl-db-secret-create:
	@echo +-----------------------------------------------------+
	@echo \| run \'make gc-db-instance-list\' to get the privat IP \|
	@echo +-----------------------------------------------------+
	kubectl create secret generic $(DATABASE) \
		--from-literal=user=postgres \
		--from-literal=pass=<PASSWD> \
		--from-literal=db=$(DATABASE) \
		--from-literal=db_host=$(PRIVATE_ADDRESS)

kctl-db-secret-delete:
	@echo run \'make gc-db-instance-list\' to get the privat IP
	@gcloud secrets delete $(DATABASE)

gc-db-instance-list:
	@gcloud sql instances list

# Create the db instance (micro, hdd, private IP, single zone)
# https://cloud.google.com/sql/pricing#pg-cpu-mem-pricing
# db-f1-micro               0.6 GB RAM, 3 GB Max Storage Cap. ($9.20/mo)
# tier: db-custom-1-3840    1.0 GB RAM, 3 GB Max Storage Cap. ($???/mo)
# gcloud beta sql instances create -h
gc-db-instance-create:
	@gcloud beta sql instances create \
		$(DATABASE_INSTANCE) \
		--async \
		--availability-type=zonal \
		--database-version=POSTGRES_12 \
		--network=default \
		--no-assign-ip \
		--root-password=${DB_ROOT_PASSWORD} \
		--storage-type=HDD \
		--tier=db-f1-micro \
		--zone=$(ZONE)

# acquire a public ip address
gc-db-instance-patch-public-ip-acquire:
	@gcloud sql instances patch $(DATABASE_INSTANCE) \
		--assign-ip \
		--async

# release the public ip address
gc-db-instance-patch-public-ip-release:
	@gcloud sql instances patch $(DATABASE_INSTANCE) \
		--no-assign-ip \
		--async

# trust the specified ip address
gc-db-instance-patch-authorized-networks:
	@gcloud sql instances patch $(DATABASE_INSTANCE) \
		--authorized-networks=$(MY_IP) \
		--async

# detrust the specified ip addresses
gc-db-instance-patch-clear-authorized-networks:
	@gcloud sql instances patch $(DATABASE_INSTANCE) \
		--clear-authorized-networks \
		--async

# https://cloud.google.com/sdk/gcloud/reference/sql/connect
gc-db-instance-connect:
	@gcloud sql connect $(DATABASE_INSTANCE) \
	 --database=$(DATABASE) \
	 --user=postgres

# 1. Enable the Cloud SQL Admin API
#    https://console.cloud.google.com/flows/enableapi?apiid=sqladmin
# 2. Download the proxy, make it executable
gc-db-cloud-proxy:
#	sudo ufw allow out 3307
	wget https://dl.google.com/cloudsql/cloud_sql_proxy.linux.amd64 -O cloud_sql_proxy
	chmod u+x cloud_sql_proxy
	sudo mkdir /cloudsql; sudo chmod 777 /cloudsql
	@echo starting \'Cloud SQL Proxy\' using \'Unix sockets\' ...
	@echo connect with \'psql \"sslmode=disable host=/cloudsql/$(PROJECT):$(REGION):$(DATABASE_INSTANCE) user=postgres\"\'
	./cloud_sql_proxy -dir=/cloudsql
#	@echo connect with \'psql "sslmode=disable host=127.0.0.1 dbname=$(DATABASE) user=postgres"\'
#	TCP sockets:
#	./cloud_sql_proxy -instances=$(PROJECT):$(REGION):$(DATABASE_INSTANCE)=tcp:5432
#	./cloud_sql_proxy \
		-instances=$(PROJECT):$(REGION):$(DATABASE_INSTANCE)=tcp:5432 \
		-credential_file=/path/to/service-account.json
#	Recommended for production environments:
#	./cloud_sql_proxy -dir=/cloudsql \
		-instances=$(PROJECT):$(REGION):$(DATABASE_INSTANCE) \
		-credential_file=/path/to/service-account.json

gc-db-instance-describe:
	@gcloud sql instances describe $(DATABASE_INSTANCE)

# https://cloud.google.com/sql/docs/postgres/start-stop-restart-instance#gcloud
# stop the db instance to save money; data is persistent.
gc-db-instance-stop:
	@gcloud sql instances patch $(DATABASE_INSTANCE) \
		--activation-policy NEVER \
		--async

# https://cloud.google.com/sql/docs/postgres/start-stop-restart-instance#gcloud
gc-db-instance-start:
	@gcloud sql instances patch $(DATABASE_INSTANCE) \
		--activation-policy ALWAYS \
		--async

# https://cloud.google.com/sql/docs/postgres/start-stop-restart-instance#gcloud
# The instance will shut down and start up again immediately:
# - if its activation policy is "always".
# - if "on demand," the instance will start up again 
#   when a new connection request is made.
gc-db-instance-restart:
	@gcloud sql instances restart $(DATABASE_INSTANCE)

gc-db-list:
	@gcloud sql databases list --instance=$(DATABASE_INSTANCE)

# https://cloud.google.com/sdk/gcloud/reference/sql/databases/create
gc-db-create:
	@gcloud beta sql databases create \
		$(DATABASE) \
		--instance=$(DATABASE_INSTANCE) \
		--async \
		--charset=UTF-8 \
		--collation=da_DK.UTF-8 \
		--verbosity=info

# https://cloud.google.com/sdk/gcloud/reference/sql/databases/patch
# does not seem to work at all; collation will stay at en_US.UTF-8 
gc-db-patch:
	@gcloud sql databases patch \
		$(DATABASE) \
		--instance=$(DATABASE_INSTANCE) \
		--charset=UTF-8 \
		--collation=da_DK.UTF-8 \
		--diff \
		--verbosity=info

# https://cloud.google.com/sdk/gcloud/reference/sql/databases/delete
gc-db-delete:
	@gcloud sql databases delete \
		$(DATABASE) \
		--instance=$(DATABASE_INSTANCE) \
		--verbosity=info

# https://cloud.google.com/sdk/gcloud/reference/sql/databases/describe
gc-db-describe:
	@gcloud sql databases describe \
		$(DATABASE) \
		--instance=$(DATABASE_INSTANCE) \
		--verbosity=info
