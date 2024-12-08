all: build up logs

up:
	mkdir -p data/gnoland/config
	cp overlay/config.toml data/gnoland/config/config.toml
	docker compose up --build -d

logs:
	docker compose logs --tail=1000 -f

down:
	docker compose down
	docker volume rm -f staginggnoland_gnonode
	docker compose run gnoland rm -rf /opt/gno/src/gno.land/gnoland-data/data /opt/gno/src/gno.land/gnoland-data/config

pull:
	git pull

build: 
	docker compose build
