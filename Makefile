.PHONY: build mk-build helm-up helm-down
export IMAGE_NAME=nyumspace
export IMAGE_TAG=local
include .env
export

HELM_CHART=deployments/helm/nyumspace
HELM_RELEASE=nyumspace

build:
	docker build -f deployments/applications/nyumspace/Dockerfile -t $(IMAGE_NAME):$(IMAGE_TAG) .

mk-build:
	eval $$(minikube docker-env) && docker build -f deployments/applications/nyumspace/Dockerfile -t nyumspace:local .

helm-up:
	minikube tunnel &
	helm upgrade --install $(HELM_RELEASE) $(HELM_CHART)

helm-down:
	helm uninstall $(HELM_RELEASE)
	pkill -f "minikube tunnel" || true


