IMAGE_NAME ?= nyumspace-api
IMAGE_TAG ?= latest
KIND            ?= kindest/node:v1.35.0
KIND_CLUSTER    ?= nyumspace-dev

.PHONY: build

build:
	docker build -f deployments/applications/nyumspace/Dockerfile -t $(IMAGE_NAME):$(IMAGE_TAG) .

deploy-dev:
	kind create cluster --name $(KIND_CLUSTER) --image $(KIND) --config=deployments/k8s/kind-config.yaml
	kind load docker-image $(IMAGE_NAME):$(IMAGE_TAG) --name $(KIND_CLUSTER)
	kubectl apply -f deployments/k8s/dev/