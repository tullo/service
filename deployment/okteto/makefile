SHELL = /bin/bash -o pipefail

export NAMESPACE = tullo
export DATABASE = postgresql
export REGISTRY_ACCOUNT = tullo
export VERSION = 1.0
export KUBECONFIG= $(HOME)/Downloads/okteto-kube.config:$(HOME)/.kube/config

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

delete:
	kubectl delete -f deploy-sales-api.yaml -f deploy-search.yaml
	kubectl delete -f zipkin-deployment.yml -f zipkin-svc.yml
	watch kubectl get pod,svc

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

shell:
	kubectl get pod -l service=sales-api
	# kubectl exec -it <POD NAME> --container sales-api  -- /bin/sh
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed

sales-api:
	okteto build \
		-f ./../../deployment/docker/dockerfile.sales-api \
		-t registry.cloud.okteto.net/$(NAMESPACE)/sales-api-amd64:$(VERSION) \
		--build-arg PACKAGE_NAME=sales-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		./../../.

metrics:
	okteto build \
		-f ./../../deployment/docker/dockerfile.metrics \
		-t registry.cloud.okteto.net/$(NAMESPACE)/metrics-amd64:$(VERSION) \
		--build-arg PACKAGE_NAME=metrics \
		--build-arg PACKAGE_PREFIX=sidecar/ \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		./../../.

port-forward:
	kubectl get pod -l service=sales-api
	#kubectl get pod --show-labels
	#kubectl port-forward sales-api-7b45cf9d76-68n8z 4040:4000

debug:
	@echo --- SALES-API
#	kubectl describe deploy/sales-api
#	kubectl describe rs/sales-api-7b45cf9d76
#	kubectl describe pod sales-api-7b45cf9d76-68n8z
#	kubectl get pod sales-api-7b45cf9d76-68n8z -o go-template="{{range .status.containerStatuses}}{{.lastState.terminated.exitCode}}{{end}}"
#	kubectl logs deploy/sales-api -c metrics | less
#	kubectl logs deploy/sales-api -c sales-api --previous | less
#	kubectl logs pod/sales-api-7b45cf9d76-68n8z -c metrics | less
#	kubectl logs pod/sales-api-7b45cf9d76-68n8z -c sales-api | less
#	kubectl logs pod/sales-api-7b45cf9d76-68n8z -c sales-api --previous | less
	@echo --- ZIPKIN
#	kubectl describe deploy/zipkin
#	kubectl describe rs/zipkin-767859464d
#	kubectl describe pod/zipkin-767859464d-m29l9
#	kubectl logs deploy/zipkin | less
	@echo --- EVENTS
#	kubectl get event -o custom-columns=NAME:.metadata.name | cut -d "." -f1