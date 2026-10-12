.PHONY: test run

GO ?= go

test:
	$(GO) test ./...

run:
	$(GO) run ./cmd/auditpattern $(AUDIT_PATTERN_FLAGS) ./expected/*.yaml
