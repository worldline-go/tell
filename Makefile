.DEFAULT_GOAL := help

.PHONY: tag
tag: ## Tags the repo with new version and tag all sub modules same time than push all
	@read -p "Enter the new version (e.g. v1.0.0): " version; \
	tags=("$$version"); \
	git tag "$$version"; \
	for dir in $(shell find . -mindepth 2 -type f -name 'go.mod' -exec dirname {} \; | sed 's|^\./||'); do \
		git tag "$$dir/$$version"; \
		tags+=("$$dir/$$version"); \
	done; \
	git tag -l | grep "$$version"; \
	read -p "Review tags above. Press enter to push to origin: "; \
	git push origin "$${tags[@]}"

.PHONY: env
env: ## Initializes a dev environment with dev dependencies
	docker compose --project-name=telemetry --file=env/compose.yaml up -d --remove-orphans

.PHONY: env-destroy
env-destroy: ## Stops the dependencies in the dev environment and destroys the data
	docker compose --project-name=telemetry down --volumes

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
