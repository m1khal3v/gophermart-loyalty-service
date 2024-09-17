help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[32m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

build: ## Build image
	docker compose build

up: ## Up containers
	docker compose up -d --remove-orphans

down: ## Down containers
	docker compose down --remove-orphans

logs: ## Show logs
	docker compose logs

logsf: ## Follow logs
	docker compose logs -f

migrate: ## Execute migrations
	docker compose run --rm goose

vet: ## Run go vet
	docker compose run --rm --no-deps service go vet ./...

test: ## Run go test
	docker compose run --rm --no-deps service go test -v ./...

test-race: ## Run go race test
	docker compose run --rm --no-deps service go test -v -race ./...

test-cover: # Run coverage
	docker compose run --rm --no-deps service bash -c "go test -v -coverpkg=./... -coverprofile=profile.cov ./... > /dev/null && go tool cover -func profile.cov"

diff: ## Generate diff migration
	docker compose -f atlas.yml build && \
	docker compose -f atlas.yml run --rm atlas migrate hash --env gorm && \
	docker compose -f atlas.yml run --rm atlas migrate diff migration --env gorm && \
	docker compose -f atlas.yml down --remove-orphans

pprof-cpu: ## Capture CPU pprof profile
	docker compose kill -s SIGUSR1 service

pprof-mem: ## Capture memory pprof profile
	docker compose kill -s SIGUSR2 service
