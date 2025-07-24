.PHONY: test clean build logs

# Run the e2e tests
test:
	@echo "Running e2e tests with Docker..."
	docker-compose build
	docker-compose up --abort-on-container-exit --exit-code-from gnokey-test

# Clean up containers and networks
clean:
	docker-compose down -v
	docker-compose rm -f

# Build containers without running tests
build:
	docker-compose build

# Show logs
logs:
	docker-compose logs -f

# Run tests in detached mode and show logs
test-detached:
	docker-compose up -d
	docker-compose logs -f gnokey-test