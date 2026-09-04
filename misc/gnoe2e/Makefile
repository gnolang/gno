rundep := go run -modfile ../devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint

# The unit lane. -short skips the scenario driver and the builder's own
# checkout build, so it answers in seconds.
GOTEST_FLAGS ?= -v

# The scenario lane: real clusters, well past the 10 minute go test default.
# Narrow a run by replacing -run, eg. -run TestScenarios/oracle/first_light.
SCENARIO_FLAGS ?= -v -timeout 30m -run TestScenarios

# The tour at the root of testdata, then every feature lane. A scenario
# declares its own cluster, so a new lane costs a directory, not a target.
# Lanes come from the scripts they hold: make 3.81 ignores the trailing slash
# in $(wildcard testdata/*/) and matches the tour as well.
SCENARIO_DIRS := testdata $(sort $(dir $(wildcard testdata/*/*.txtar)))

.PHONY: install
install:
	go install .

.PHONY: build
build:
	go build -o build/gnoe2e .

.PHONY: lint
lint:
	$(golangci_lint) --config ../../.github/golangci.yml run ./...

.PHONY: test
test:
	go test $(GOTEST_FLAGS) -short ./...

# The scenario driver is a test of the root package, hence "." rather than
# ./..., which would boot a cluster and then run the unit tests again.
.PHONY: test-scenarios
test-scenarios:
	go test $(SCENARIO_FLAGS) .

.PHONY: test-all
test-all:
	go test $(GOTEST_FLAGS) -timeout 30m ./...

# test-master runs the scenarios against gnoland and gpao built from master,
# so a scenario written for a fix can be shown red before the fix lands. The
# ref is unpacked with git archive into a temp dir the trap removes, so a
# Ctrl-C leaves nothing registered in .git.
.PHONY: test-master
test-master:
	tmp="$$(mktemp -d)" && trap 'rm -rf "$$tmp"' EXIT && \
	git -C ../.. archive master | tar -x -C "$$tmp" && \
	GNOROOT="$$tmp" go test $(SCENARIO_FLAGS) .

# The CLI route: coloured output, and every scenario runs even after one
# fails. Use "test-scenarios" for the go test route.
.PHONY: scenarios
scenarios: build
	build/gnoe2e run -verbose $(SCENARIO_DIRS)
