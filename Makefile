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

diff:
	docker compose -f atlas.yml build && \
	docker compose -f atlas.yml run --rm atlas migrate hash --env gorm && \
	docker compose -f atlas.yml run --rm atlas migrate diff --env gorm && \
	docker compose -f atlas.yml down --remove-orphans
