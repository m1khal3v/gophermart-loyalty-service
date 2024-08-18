build:
	docker compose build

up:
	docker compose up -d --remove-orphans

down:
	docker compose down --remove-orphans

logs:
	docker compose logs

migrate:
	docker compose run --rm goose

vet:
	docker compose run --rm --no-deps service go vet ./...

test:
	docker compose run --rm --no-deps service go test -v ./...

test-race:
	docker compose run --rm --no-deps service go test -v -race ./...

diff:
	docker compose -f atlas.yml build && \
	docker compose -f atlas.yml run --rm atlas migrate hash --env gorm && \
	docker compose -f atlas.yml run --rm atlas migrate diff migration --env gorm && \
	docker compose -f atlas.yml down --remove-orphans
