# Gno-vs-Go differential fuzzer.
#
# `test` is the fast suite (comparator, loader, one gc/GnoVM smoke program).
# `corpus` runs every seed through both toolchains and is intentionally not
# part of the default `make test` / `go test` path.

.PHONY: test corpus run-generated fuzz

test:
	go test ./...

corpus:
	go run . -out gnodiff-out

run-generated:
	go run . -generate 2000 -seed 1 -timeout 30s -out gnodiff-out/generated

fuzz:
	go test -fuzz=FuzzGenerated -fuzztime=30s ./internal/diff/
