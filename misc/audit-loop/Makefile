.PHONY: test run

test:
	go test ./...

run:
	go run ./cmd/auditloop $(AUDIT_LOOP_FLAGS) ./expected/*.yaml
