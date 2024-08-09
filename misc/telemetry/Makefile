.PHONY: up
up:
ifeq ($(filter fresh, $(MAKECMDGOALS)), fresh)
	$(MAKE) build-local-image-no-cache
else
	$(MAKE) build-local-image
endif
	docker compose up -d --build

.PHONY: down
down:
	docker compose down

.PHONY: clean
clean:
	docker compose down -v

.PHONY: build-local-image
build-local-image:
	docker build -f ../../Dockerfile.local -t local-gnoland:latest ../../

.PHONY: build-local-image-no-cache
build-local-image-no-cache:
	docker build --no-cache -f ../../Dockerfile.local -t local-gnoland:latest ../../
