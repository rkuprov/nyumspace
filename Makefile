.PHONY: build mk-build mk-up mk-down mk-run
export IMAGE_NAME=nyumspace
export IMAGE_TAG=local
include .env
export

build:
	docker build -f deployments/applications/nyumspace/Dockerfile -t $(IMAGE_NAME):$(IMAGE_TAG) .

mk-build:
	eval $$(minikube docker-env) && docker build -f deployments/applications/nyumspace/Dockerfile -t nyumspace:local .

mk-up:
	kubectl apply -f deployments/k8s/nyumspace/namespace.yaml
	kubectl apply -f deployments/k8s/nyumspace/deployment.yaml
	kubectl apply -f deployments/k8s/nyumspace/service.yaml

mk-down:
	kubectl delete -f deployments/k8s/nyumspace/service.yaml
	kubectl delete -f deployments/k8s/nyumspace/deployment.yaml
	kubectl delete -f deployments/k8s/nyumspace/namespace.yaml

mk-run:
	kubectl port-forward -n nyumspace svc/nyumspace 8000:8080


