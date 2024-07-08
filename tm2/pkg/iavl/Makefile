GOTOOLS := github.com/golangci/golangci-lint/cmd/golangci-lint

PDFFLAGS := -pdf --nodefraction=0.1

LDFLAGS := -ldflags "-X TENDERMINT_IAVL_COLORS_ON=on"

all: lint test install



install:
ifeq ($(COLORS_ON),)
	go install ./cmd/iaviewer
else
	go install $(LDFLAGS) ./cmd/iaviewer
endif

test:
	go test -v --race

tools:
	go get -v $(GOTOOLS)

# look into .golangci.yml for enabling / disabling linters
lint:
	@echo "--> Running linter"
	@golangci-lint run

# bench is the basic tests that shouldn't crash an aws instance
bench:
	cd benchmarks && \
		go test -bench=RandomBytes . && \
		go test -bench=Small . && \
		go test -bench=Medium . && \
		go test -bench=BenchmarkMemKeySizes .

# fullbench is extra tests needing lots of memory and to run locally
fullbench:
	cd benchmarks && \
		go test -bench=RandomBytes . && \
		go test -bench=Small . && \
		go test -bench=Medium . && \
		go test -timeout=30m -bench=Large . && \
		go test -bench=Mem . && \
		go test -timeout=60m -bench=LevelDB .


# note that this just profiles the in-memory version, not persistence
profile:
	cd benchmarks && \
		go test -bench=Mem -cpuprofile=cpu.out -memprofile=mem.out . && \
		go tool pprof ${PDFFLAGS} benchmarks.test cpu.out > cpu.pdf && \
		go tool pprof --alloc_space ${PDFFLAGS} benchmarks.test mem.out > mem_space.pdf && \
		go tool pprof --alloc_objects ${PDFFLAGS} benchmarks.test mem.out > mem_obj.pdf

explorecpu:
	cd benchmarks && \
		go tool pprof benchmarks.test cpu.out

exploremem:
	cd benchmarks && \
		go tool pprof --alloc_objects benchmarks.test mem.out

delve:
	dlv test ./benchmarks -- -test.bench=.

.PHONY: lint test tools install delve exploremem explorecpu profile fullbench bench
