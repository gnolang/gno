.PHONY: test clean logs

test:
	@echo "Running e2e tests with Docker..."
	docker compose build
	docker compose up --abort-on-container-exit --exit-code-from gnokey-test; \
	EXIT=$$?; \
	echo ""; \
	echo "========== TEST REPORT =========="; \
	docker compose logs gnokey-test 2>/dev/null | grep -E "\[PASS\]|\[FAIL\]|\[KNOWN\]|PASS:|FAIL:|KNOWN:|All tests|Some tests"; \
	echo "================================="; \
	exit $$EXIT

clean:
	docker compose down -v
	docker compose rm -f

logs:
	docker compose logs -f
