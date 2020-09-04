#!make
SHELL = /bin/bash -o pipefail
VERSION = 1.0
ACCOUNT = xyz@gmail.com
BILLING_ACCOUNT_ID = ABCDEF-GHIJKL-NMOPQR
SERVICE_ACCOUNT = stackwise
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
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
SOPS_PGP_FP='gpg key fingerprint list; comma separated'
include .env
#export

.DEFAULT_GOAL := show-configs

#==============================================================================
# GKE Installation
#
# Install the Google Cloud SDK. 
# This contains the gcloud client needed to perform sdk related tasks
# https://cloud.google.com/sdk/
#
# Install the K8s kubectl client. 
# https://kubernetes.io/docs/tasks/tools/install-kubectl/
#
#
# A region is a specific geographical location where you can host your resources.
# Each region has one or more zones.
# Resources that live in a zone, such as virtual machine instances or zonal persistent disks,
# are referred to as zonal resources.
# Other resources, like static external IP addresses, are regional.
# https://cloud.google.com/compute/docs/regions-zones
#
#
# Google designs zones to be independent from each other:
# - a zone usually has power, cooling, networking, and control planes that are isolated from other zones
# - most single failure events will affect only a single zone.
# - https://cloud.google.com/compute/docs/regions-zones#available
# - europe-west3-c (region:europe-west3 zone:c)
#
# Generally, communication within regions will always be cheaper and faster than communication across different regions.
# To mitigate effects of possible events, you should duplicate important systems in multiple zones and regions.
# eg. europe-west3-a, europe-west3-b, etc
# https://cloud.google.com/compute/docs/regions-zones#choosing_a_region_and_zone
#

show-configs: kctl-config-contexts gc-configurations-list

# https://github.com/mozilla/sops
sops-encrypt-file-using-sops.yaml:
	@echo ====== encrypt secrets using keys defined in \.sops.yaml ====
	@$(shell go env GOPATH)/bin/sops config.staging.yaml

sops-encrypt-secrets:
	@echo ====== encrypt secrets using fingerpring:  ${SOPS_PGP_FP} ====
	@$(shell go env GOPATH)/bin/sops --verbose -p ${SOPS_PGP_FP} -e .env > .enc.env
	@$(shell go env GOPATH)/bin/sops --verbose -p ${SOPS_PGP_FP} -e .env-db > .enc.env-db

sops-decrypt-secrets:
	@echo ====== decrypt secrets ====
	@$(shell go env GOPATH)/bin/sops --verbose -d .enc.env > .env
	@$(shell go env GOPATH)/bin/sops --verbose -d .enc.env-db > .env-db

sops-edit-env:
	@$(shell go env GOPATH)/bin/sops --verbose -p ${SOPS_PGP_FP} .enc.env

sops-edit-env-db:
	@$(shell go env GOPATH)/bin/sops --verbose -p ${SOPS_PGP_FP} .enc.env-db

sops-add-access:
	gpg --list-keys
	@$(shell go env GOPATH)/bin/sops --verbose --rotate --in-place --add-pgp <key-id> .enc.env-db

sops-remove-access:
	gpg --list-keys
	@$(shell go env GOPATH)/bin/sops --verbose --rotate --in-place --rm-pgp <key-id> .enc.env-db

# https://github.com/mozilla/sops#adding-and-removing-keys
sops-add-access-with-sops.yaml:
	gpg --list-keys
	@echo =======================================+
	@echo '==>' add key to the \.sops.yaml file.
	@echo =======================================+
	@cat ../../../.sops.yaml
	@$(shell go env GOPATH)/bin/sops updatekeys .enc.env-db

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

#==============================================================================
# Cloud SDK goals                                                             #
#==============================================================================

gc-projects-create:
	@gcloud projects create $(PROJECT)

gc-init: gc-projects-create
	@gcloud init

gc-config-list:
	@gcloud config list

gc-config-set:
	@echo '==>' setting project: ${PROJECT} and compute zone: $(ZONE)
	@gcloud config set project ${PROJECT}
	@gcloud config set compute/zone $(ZONE)

config: gc-projects-create gc-config-set gc-auth-configure-docker
	@gcloud beta billing projects link $(PROJECT) --billing-account=$(BILLING_ACCOUNT_ID)

# https://cloud.google.com/sdk/docs/configurations
gc-configurations-list:
	@echo ============================================================================+
	@echo '==>' gcloud config configurations list - lists existing named configurations \|
	@echo ============================================================================+
	gcloud config configurations list

gc-configurations-describe:
	@gcloud config configurations describe cloud-experiments

gc-configurations-delete:
	@gcloud config configurations delete <config>

gc-config-set-account: gc-auth-list
	@gcloud config set account ${ACCOUNT}

# listing of enabled services.
gc-svc-list:
	@gcloud services list

# enable cloud sql admin api.
gc-svc-enable-sqladmin:
	@gcloud services enable sqladmin.googleapis.com

gc-auth-list:
	@echo '==>' gcloud auth list - lists credentialed accounts.
	@gcloud auth list

# obtain access credentials for your user account via a web-based authorization flow.
gc-auth-login:
	@echo '==>' Authorize with a user account without setting up a configuration.
	@gcloud auth login

# same function as [gcloud auth login] but uses a service account.
# gcloud auth activate-service-account -h
gc-auth-activate-service-account:
	@echo '==>' Authorize with a service account instead of a user account.
	@echo '==>' Open in browser: 'https://console.cloud.google.com/iam-admin/serviceaccounts'
	@echo '==>' Create service account with ID: \'${SERVICE_ACCOUNT}\'
	@echo '==>' From the \'Role\' list, select \'Project \> Owner\'.
	@echo '==>' From the \'Actions\' list, select \'Create key\', choose \'json\' format.
	@gcloud auth activate-service-account --key-file=/path/to/service-account.json

# list existing clusters.
gc-clusters-list:
	@gcloud container clusters list

# fetch credentials for a running cluster.
gc-clusters-get-credentials:
	@gcloud container clusters get-credentials $(CLUSTER) \
		--project $(PROJECT) \
		--zone $(ZONE)


# https://cloud.google.com/container-registry/docs/pushing-and-pulling
# register gcloud as the credential helper for the Google-supported Docker registry.
gc-auth-configure-docker:
	gcloud auth configure-docker $(CONTAINER_REGISTRY)

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

#==============================================================================
# Docker goals                                                                #
#==============================================================================

docker-tag-images:
	@docker tag tullo/sales-api-amd64:$(VERSION) $(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64:$(VERSION)
	@docker tag tullo/metrics-amd64:$(VERSION) $(CONTAINER_REGISTRY)/$(PROJECT)/metrics-amd64:$(VERSION)

docker-push-images:
	@docker image push $(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64:$(VERSION)
	@docker image push $(CONTAINER_REGISTRY)/$(PROJECT)/metrics-amd64:$(VERSION)

#==============================================================================
# kubectl goals                                                               #
#==============================================================================

# validate kubernetes yaml manifests
kubeval:
	@kustomize build . | kubeval --strict --force-color -

# modify kustomization.yaml and add a SecretGenerator.
deployment-add-secret:
	@kustomize edit add secret $(DATABASE) --from-env-file=.env-db
	@echo ==== add secret generator ====
	@echo secretGenerator:
	@echo - envs:
	@echo   - .env-db
	@echo   name: $(DATABASE)
	@echo   type: Opaque

# change an image of the deployment.
# note: the image tag HASH/VERSION is usualy defined by a CI/CD system.
deployment-set-image:
	@echo ==============================================================================+
	@echo '==>' kustomize edit set image - Make sure the image is pullable on the cluster \|
	@echo ==============================================================================+
	@$(eval HASH=`git rev-parse HEAD`)
	@kustomize edit set image $(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64=$(CONTAINER_REGISTRY)/$(PROJECT)/sales-api-amd64:$(HASH)

deployment-apply: kubeval
	@kustomize build . | kubectl apply --validate -f -
#	@kustomize build . | kubectl apply --dry-run=client --validate -f -
	@watch kubectl get po

deployment-delete:
	kustomize build . | kubectl delete -f -

kctl-config-contexts:
	@echo ================================================================================+
	@echo '==>' kubectl config get-contexts - List all the contexts in your kubeconfig file \|
	@echo ================================================================================+
	@kubectl config get-contexts

kctl-config-use-context:
	@kubectl config use-context gke_$(PROJECT)_$(ZONE)_$(CLUSTER)

kctl-cluster-info:
	@kubectl cluster-info --context gke_$(PROJECT)_$(ZONE)_$(CLUSTER)

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

#==============================================================================
# APP goals                                                                   #
#==============================================================================

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

#==============================================================================
# DB instance goals                                                           #
#==============================================================================

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

gc-db-instance-delete:
	gcloud sql instances delete $(DATABASE_INSTANCE) --async

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
# firewalls eventually block outgoing port 3307
# sudo ufw allow out 3307
gc-db-cloud-proxy:
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

#==============================================================================
# Database goals                                                              #
#==============================================================================

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

# see Notes.md: [Create a Kubernetes Cluster]
cluster: gc-clusters-list
	@gcloud compute instances list
#	@gcloud compute machine-types list | grep ${ZONE}
	@$(eval k8s_version=`gcloud container get-server-config --zone=${ZONE} --format=json | jq -r '.validNodeVersions[0]'`)
	@echo Current valid node sersion: ${k8s_version}
	@gcloud container clusters describe $(CLUSTER) --zone=${ZONE}
#	@gcloud container clusters create $(CLUSTER) \
		--cluster-version=${k8s_version} \
		--zone=${ZONE} \
		--num-nodes=1 \
		--machine-type=e2-small \
		--disk-size=30 \
		--enable-ip-alias \
		--enable-autoupgrade
