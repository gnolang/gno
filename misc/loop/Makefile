all: docker.start

docker.start: # Start the portal loop
	docker compose up -d

docker.stop: # Stop the portal loop
	docker compose down
	docker rm -f $(docker ps -aq --filter "label=the-portal-loop")

docker.build: # (re)Build snapshotter image
	docker compose build

docker.pull: # Pull new images to update versions
	docker compose pull

portalloopd.bash: # Get a bash command inside of the portalloopd container
	docker compose exec portalloopd bash

switch: portalloopd.switch

portalloopd.switch: # Force switch the portal loop with latest image
	docker compose exec portalloopd switch
