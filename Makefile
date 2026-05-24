.PHONY: build mk-build helm-up helm-down
.PHONY: _migrate-local migrate-up migrate-down migrate-status
.PHONY: migrate-up-sandbox migrate-down-sandbox migrate-up-prod migrate-down-prod
export IMAGE_NAME=nyumspace
export IMAGE_TAG=local
include .envrc
export

HELM_CHART=deployments/helm/nyumspace
HELM_RELEASE=nyumspace

# Migration settings — override on the command line to target a different cluster/namespace
PG_NAMESPACE   ?= nyumspace
PG_USER        ?= admin
PG_PASSWORD    ?= pgpassword
PG_DB          ?= nyumdb
PG_LOCAL_PORT  ?= 5433
MIGRATIONS_DIR ?= deployments/migrations

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

# Internal: opens a port-forward to the in-cluster postgres, runs goose, then tears it down.
# Call via the public targets below; do not invoke directly.
_migrate-local:
	@kubectl port-forward -n $(PG_NAMESPACE) svc/postgres $(PG_LOCAL_PORT):5432 & \
	PF_PID=$$!; \
	sleep 2; \
	GOOSE_DRIVER=postgres \
	GOOSE_DBSTRING="postgres://$(PG_USER):$(PG_PASSWORD)@localhost:$(PG_LOCAL_PORT)/$(PG_DB)?sslmode=disable" \
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_CMD); \
	RESULT=$$?; \
	kill $$PF_PID 2>/dev/null; \
	exit $$RESULT

migrate-up: ## Apply all pending migrations (local k8s)
	@$(MAKE) _migrate-local GOOSE_CMD=up

migrate-down: ## Roll back the last migration (local k8s)
	@$(MAKE) _migrate-local GOOSE_CMD=down

migrate-status: ## Show migration status (local k8s)
	@$(MAKE) _migrate-local GOOSE_CMD=status

migrate-up-sandbox: ## Apply migrations to sandbox (not yet implemented)
	@echo "sandbox migrations not yet implemented"

migrate-down-sandbox: ## Roll back migration on sandbox (not yet implemented)
	@echo "sandbox migrations not yet implemented"

migrate-up-prod: ## Apply migrations to prod (not yet implemented)
	@echo "prod migrations not yet implemented"

migrate-down-prod: ## Roll back migration on prod (not yet implemented)
	@echo "prod migrations not yet implemented"
