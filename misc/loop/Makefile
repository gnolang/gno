PULL_GH_SCRIPT := ./scripts/pull-gh.sh

all: docker.start.prod

docker.start.prod: # Start the production portal loop
	docker compose -f docker-compose.yml -f docker-compose.override.prod.yml up -d

docker.start: # Start the portal loop
	docker compose up -d

docker.ci: # Start the portal loop for CI
	docker compose up -d portalloopd traefik

docker.stop: # Stop the portal loop
	docker compose down
	docker rm -f $(docker ps -aq --filter "label=the-portal-loop")

portalloopd.bash: # Get a bash command inside of the portalloopd container
	docker compose exec portalloopd bash

switch: portalloopd.switch

portalloopd.switch: # Force switch the portal loop with latest image
	docker compose exec portalloopd switch

prepare-exports:
	chmod +x $(PULL_GH_SCRIPT) && ./$(PULL_GH_SCRIPT)

pull-exports: docker.stop prepare-exports
	docker.start.prod

