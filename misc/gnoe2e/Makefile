rundep := go run -modfile ../devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint

# The unit lane. -short skips the scenario driver and the builder's own
# checkout build, so it answers in seconds -- cheaply enough to carry the race
# detector, which is the only thing that reads the harness's own goroutines. The
# scenario lane cannot: it exercises them through child processes, where the
# detector sees nothing.
GOTEST_FLAGS ?= -v -race

# The scenario lane: real clusters, well past the 10 minute go test default.
# Narrow a run by replacing -run, eg. -run TestScenarios/oracle/first_light.
SCENARIO_FLAGS ?= -v -timeout 30m -run TestScenarios

# The tour at the root of testdata, then every feature lane. A scenario
# declares its own cluster, so a new lane costs a directory, not a target.
# Lanes come from the scripts they hold: make 3.81 ignores the trailing slash
# in $(wildcard testdata/*/) and matches the tour as well.
SCENARIO_DIRS := testdata $(sort $(dir $(wildcard testdata/*/*.txtar)))

.DEFAULT_GOAL := all

.PHONY: all
all: build ## Build the gnoe2e binary

.PHONY: help
help: ## List the targets
	grep -hE '^[a-z][a-z-]*:.*?## ' $(MAKEFILE_LIST) | sort | \
	awk -F ':.*?## ' '{printf "%-16s %s\n", $$1, $$2}'

.PHONY: install
install: ## Install gnoe2e into GOBIN
	go install .

.PHONY: build
build: ## Build build/gnoe2e
	go build -o build/gnoe2e .

.PHONY: clean
clean: ## Remove what build produced
	rm -rf build

.PHONY: lint
lint: ## Run the repository's linter on this module
	$(golangci_lint) --config ../../.github/golangci.yml run ./...

# Every "-- cluster --" key a scenario can declare, with the value a cluster
# boots with. The values are the harness's, not the chain's.
.PHONY: defaults
defaults: ## Print every "-- cluster --" key with its default
	go run . defaults

.PHONY: test
test: ## Run the unit lane, in seconds, booting nothing
	go test $(GOTEST_FLAGS) -short ./...

# The scenario driver is a test of the root package, hence "." rather than
# ./..., which would boot a cluster and then run the unit tests again.
.PHONY: test-scenarios
test-scenarios: ## Run every scenario against real clusters
	go test $(SCENARIO_FLAGS) .

# Both lanes, one after the other rather than as one ./... run: each has flags
# of its own, and running them together lets the builder's real gnoland build
# compete with the clusters the scenarios are booting.
.PHONY: test-all
test-all: test test-scenarios ## Run both lanes

# test-master runs the scenarios against gnoland and gpao built from master,
# so a scenario written for a fix can be shown red before the fix lands. The
# ref is unpacked with git archive into a temp dir the trap removes, so a
# Ctrl-C leaves nothing registered in .git.
# The ref is checked before it is archived, and the unpacked tree before it is
# used: tar reports the exit status of a pipeline whose left-hand side failed as
# its own success, and GNOROOT pointing at an empty directory would run the
# scenarios against a checkout that is not there.
.PHONY: test-master
test-master: ## Run the scenarios against gnoland and gpao built from master
	tmp="$$(mktemp -d)" && trap 'rm -rf "$$tmp"' EXIT && \
	git -C ../.. rev-parse --verify -q master >/dev/null && \
	git -C ../.. archive master | tar -x -C "$$tmp" && \
	test -f "$$tmp/go.mod" && \
	GNOROOT="$$tmp" go test $(SCENARIO_FLAGS) .

# The CLI route: coloured output, and every scenario runs even after one
# fails. Use "test-scenarios" for the go test route.
.PHONY: scenarios
scenarios: build ## Run every scenario through the CLI, coloured
	build/gnoe2e run -verbose $(SCENARIO_DIRS)
