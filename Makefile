.DEFAULT_GOAL := help

.PHONY: env
env: ## Initializes a dev environment with dev dependencies
	docker compose --project-name=telemetry --file=env/compose.yaml up -d --remove-orphans

.PHONY: env-destroy
env-destroy: ## Stops the dependencies in the dev environment and destroys the data
	docker compose --project-name=telemetry down --volumes

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
