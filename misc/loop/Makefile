all: docker.start

docker.start: # Start the portal loop
	docker compose up -d

docker.stop: # Stop the portal loop
	docker compose down

docker.build: # (re)Build snapshotter image
	docker compose build

docker.pull: # Pull new images to update versions
	docker compose pull


snapshotter.bash: # Get a bash command inside of the snapshotter container
	docker compose exec snapshotter bash

snapshotter.switch-now: # Force switch the portal loop with latest image
	docker compose exec snapshotter /scripts/switch.sh
