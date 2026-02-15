.PHONY: build
include .env
export

build:
	docker build -f deployments/applications/nyumspace/Dockerfile -t $(IMAGE_NAME):$(IMAGE_TAG) .

up:
	docker-compose -f deployments/dev/compose.yaml up -d
	sleep 5
	temporal operator namespace create -n nyumspace

down:
	docker-compose -f deployments/dev/compose.yaml down

run:
	go build -o bin/nyumspace-api applications/nyumspace/api/main.go
	./bin/nyumspace-api

workers:
	go build -o bin/nyumspace-workers applications/nyumspace/api/workers/main.go
	./bin/nyumspace-workers