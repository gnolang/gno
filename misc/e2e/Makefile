.PHONY: test clean logs

test:
	@echo "Running e2e tests with Docker..."
	docker compose build
	docker compose up --abort-on-container-exit --exit-code-from gnokey-test

clean:
	docker compose down -v
	docker compose rm -f

logs:
	docker compose logs -f
