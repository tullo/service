SHELL := /bin/bash

NAMESPACE = tullo
DATABASE = postgresql
REGISTRY_ACCOUNT = tullo
VERSION = 1.0
KUBECONFIG= $(HOME)/Downloads/okteto-kube.config:$(HOME)/.kube/config

config:
	@echo Setting environment for $(NAMESPACE)
#	switch namespace
	okteto namespace $(NAMESPACE)
	@echo ======================================================================

diff:
	kubectl diff -f deploy-sales-api.yaml

dry-run:
	kubectl apply --dry-run -f deploy-sales-api.yaml -o yaml

deployment:
	kubectl apply -f deploy-sales-api.yaml -f deploy-search.yaml
	kubectl apply -f zipkin-deployment.yml -f zipkin-svc.yml
	@echo
	watch kubectl get pod,svc
#	kubectl logs -f deployment/sales-api --container sales-api
	@echo ======================================================================

status:
	kubectl config get-clusters
	@echo
	kubectl config current-context
	@echo
#	kubectl get pod -o wide
	kubectl get pod
	@echo
#	kubectl get svc
#	kubectl get all -l app=postgresql --show-labels
	@echo
	kubectl get svc/sales-api -o wide
	@echo ======================================================================

shell:
	kubectl get pod -l service=sales-api
	# kubectl exec -it <POD NAME> --container sales-api  -- /bin/sh
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed
	@echo ======================================================================

sales-api:
	cd .. && \
	okteto build \
		-f dockerfile.sales-api \
		-t registry.cloud.okteto.net/tullo/sales-api-amd64:$(VERSION) \
		--build-arg PACKAGE_NAME=sales-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

search:
	cd .. && \
	okteto build \
		-f dockerfile.search \
		-t registry.cloud.okteto.net/tullo/search-amd64:$(VERSION) \
		--build-arg PACKAGE_NAME=search \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

metrics:
	cd .. && \
	okteto build \
		-f dockerfile.metrics \
		-t registry.cloud.okteto.net/tullo/metrics-amd64:$(VERSION) \
		--build-arg PACKAGE_NAME=metrics \
		--build-arg PACKAGE_PREFIX=sidecar/ \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

delete:
	kubectl delete -f deploy-sales-api.yaml
	docker image remove $(REGISTRY_ACCOUNT)/sales-api-amd64:$(VERSION)
	docker image remove $(REGISTRY_ACCOUNT)/metrics-amd64:$(VERSION)
	@echo ======================================================================
