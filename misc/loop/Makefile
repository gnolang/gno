PULL_GH_SCRIPT := ./scripts/pull-gh.sh

all: docker.start.prod

clean: docker.rm

docker.start.prod: # Start the production portal loop
	docker compose -f docker-compose.yml -f docker-compose.override.prod.yml up -d

docker.start: # Start the portal loop
	docker compose up -d

docker.ci: # Start the portal loop for CI
	docker compose up -d portalloopd traefik

docker.rm: # Stop the portal loop
	docker compose stop
	docker compose down -v

switch: portalloopd.switch

portalloopd.switch: # Force switch the portal loop with latest image
	docker compose exec portalloopd portalloopd switch

prepare-exports:
	chmod +x $(PULL_GH_SCRIPT) && ./$(PULL_GH_SCRIPT)

pull-exports: docker.stop prepare-exports
	docker.start.prod

