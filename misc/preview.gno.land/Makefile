all: up logs

up:
	mkdir -p data/gnoland/config
	cp overlay/config.toml data/gnoland/config/config.toml
	docker compose up --build -d

logs:
	docker compose logs --tail=1000 -f

down:
	docker compose down

