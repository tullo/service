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
	kubectl apply -f deploy-sales-api.yaml
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

delete:
	kubectl delete -f deploy-sales-api.yaml
	docker image remove $(REGISTRY_ACCOUNT)/sales-api-amd64:$(VERSION)
	docker image remove $(REGISTRY_ACCOUNT)/metrics-amd64:$(VERSION)
	@echo ======================================================================
