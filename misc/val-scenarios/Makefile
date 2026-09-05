WORK_ROOT         ?= /tmp/gno-val-tests
NODE              ?=
LOCAL_GNO_ROOT    := $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/../..)

# RUNTIME selects the backend (matches the default in lib/scenario.sh):
#   docker (default): nodes and CLI run in containers; `make build` builds images.
#   local:            nodes and CLI run from native binaries; `make build` builds
#                     binaries into BIN_DIR.
# It is passed through to the scenario scripts, so the usage path is identical
# for both backends — only the value of RUNTIME changes:
#   make build && make scenario-01
#   make build RUNTIME=local && make scenario-01 RUNTIME=local
RUNTIME           ?= docker

# Base dir for the local-runtime binaries (RUNTIME=local). The effective BIN_DIR
# is derived per-source below: <base>/bin for a local build (matches the BIN_DIR
# default in lib/scenario.sh) or <base>/bin/<tag> for a GH_USER fork build, so
# fork and local binaries coexist the same way tagged images do.
_BIN_BASE         := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

# Remote fork support: set GH_USER to build from a GitHub fork. This drives both
# `build-images` and `build-binaries` — the fork is cloned once and used as the
# source root for every artifact.
# GH_REPO defaults to "gno" and GH_BRANCH defaults to "master".
# e.g. make build-images   GH_USER=gnolang
#      make build-binaries GH_USER=gnolang GH_BRANCH=feat/my-branch
# Image tags and the binary output dir are derived automatically from
# <GH_USER>-<GH_BRANCH> (slashes in the branch name are replaced with dashes).
GH_USER           ?=
GH_REPO           ?= gno
GH_BRANCH         ?= master

_REMOTE_GNO_ROOT  := /tmp/gno-remote-build
_FETCH_STAMP      := $(_REMOTE_GNO_ROOT)/.fetched

ifneq ($(GH_USER),)
_IMAGE_TAG        := $(GH_USER)-$(subst /,-,$(GH_BRANCH))
IMAGE             ?= gno-val-scenario-core:$(_IMAGE_TAG)
GNOKEY_IMAGE      ?= $(IMAGE)
GNOGENESIS_IMAGE  ?= gnogenesis:$(_IMAGE_TAG)
VALSIGNER_IMAGE   ?= valsignerd:$(_IMAGE_TAG)
GNO_ROOT          := $(_REMOTE_GNO_ROOT)
CORE_GNO_ROOT     ?= $(GNO_ROOT)
GNOGENESIS_GNO_ROOT ?= $(GNO_ROOT)
VALSIGNER_GNO_ROOT ?= $(GNO_ROOT)
BIN_DIR           ?= $(_BIN_BASE)/bin/$(_IMAGE_TAG)
_FETCH_PREREQ     := $(_FETCH_STAMP)
else
IMAGE             ?= gno-val-scenario-core:local
GNOKEY_IMAGE      ?= $(IMAGE)
GNOGENESIS_IMAGE  ?= gnogenesis:local
VALSIGNER_IMAGE   ?= valsignerd:local
GNO_ROOT          ?= $(LOCAL_GNO_ROOT)
CORE_GNO_ROOT     ?= $(GNO_ROOT)
GNOGENESIS_GNO_ROOT ?= $(GNO_ROOT)
VALSIGNER_GNO_ROOT ?= $(GNO_ROOT)
BIN_DIR           ?= $(_BIN_BASE)/bin
_FETCH_PREREQ     :=
endif

# Auto-discover scenario scripts, ordered by filename.
SCENARIO_SCRIPTS := $(sort $(wildcard scenarios/*.sh))
CI_SCENARIO_SCRIPTS := $(shell grep -l '^SCENARIO_CI=true$$' $(SCENARIO_SCRIPTS) 2>/dev/null | sort)
LOCAL_ONLY_SCENARIO_SCRIPTS := $(shell grep -l '^SCENARIO_CI=false$$' $(SCENARIO_SCRIPTS) 2>/dev/null | sort)

# Extract the numeric prefix from a script path.
# e.g. scenarios/02_three_validators_restart_staggered.sh -> 02
scenario_num = $(word 1,$(subst _, ,$(notdir $(patsubst %.sh,%,$(1)))))

# Generate scenario-NN / logs-NN / clean-NN rules for every discovered script.
# The docker-compose project name matches the scenario_init argument: scenario-NN.
define SCENARIO_RULES
.PHONY: scenario-$(call scenario_num,$(1)) logs-$(call scenario_num,$(1)) clean-$(call scenario_num,$(1))

scenario-$(call scenario_num,$(1)):
	RUNTIME=$(RUNTIME) IMAGE_NAME=$(IMAGE) GNOKEY_IMAGE=$(GNOKEY_IMAGE) GNOGENESIS_IMAGE=$(GNOGENESIS_IMAGE) VALSIGNER_IMAGE=$(VALSIGNER_IMAGE) GNO_ROOT=$(GNO_ROOT) BIN_DIR=$(BIN_DIR) WORK_ROOT=$(WORK_ROOT) ./$(1)

logs-$(call scenario_num,$(1)):
ifeq ($(RUNTIME),local)
	tail -f $(WORK_ROOT)/scenario-$(call scenario_num,$(1))/logs/$(if $(NODE),$(NODE),*).log
else
	docker compose -p scenario-$(call scenario_num,$(1)) logs -f $(NODE)
endif

clean-$(call scenario_num,$(1)):
	rm -rf $(WORK_ROOT)/scenario-$(call scenario_num,$(1))

endef

$(foreach s,$(SCENARIO_SCRIPTS),$(eval $(call SCENARIO_RULES,$(s))))

# Scenario target groups.
SCENARIO_TARGETS := $(foreach s,$(SCENARIO_SCRIPTS),scenario-$(call scenario_num,$(s)))
CI_SCENARIO_TARGETS := $(foreach s,$(CI_SCENARIO_SCRIPTS),scenario-$(call scenario_num,$(s)))
LOCAL_ONLY_SCENARIO_TARGETS := $(foreach s,$(LOCAL_ONLY_SCENARIO_SCRIPTS),scenario-$(call scenario_num,$(s)))

.PHONY: build test test-ci ci test-local local-only test-all test-basics basics test-advanced advanced build-core-image build-gnoland-image build-gnokey-image build-gnogenesis-image build-valsigner-image build-images build-core-binaries build-gnoland-binary build-gnokey-binary build-gnogenesis-binary build-valsigner-binary build-binaries fetch-remote clone-gno clean status

# Single build entry point: builds whatever the selected RUNTIME needs.
# `make build` builds Docker images; `make build RUNTIME=local` builds binaries.
ifeq ($(RUNTIME),local)
build: build-binaries
else
build: build-images
endif

# `.local` suffix shorthand: `make <target>.local` runs <target> with
# RUNTIME=local. Works for every target (build.local, test.local,
# scenario-01.local, logs-05.local, ...) via the stem $*, so it covers current
# and future targets without duplication. Extra command-line variables
# propagate to the sub-make, e.g. `make scenario-12.local GH_USER=gnolang`.
# Docker is the default, so there is no `.docker` twin: `make <target>` is it.
%.local:
	@$(MAKE) $* RUNTIME=local

# Run the CI scenarios in order, stopping on the first failure.
test: test-ci

test-ci ci test-basics basics: $(CI_SCENARIO_TARGETS)

# Run the locally-oriented scenarios in order.
test-local local-only test-advanced advanced: $(LOCAL_ONLY_SCENARIO_TARGETS)

# Run every scenario in order.
test-all: $(SCENARIO_TARGETS)

# Stamp-file rule: clones once; subsequent builds skip it until fetch-remote is called.
$(_FETCH_STAMP):
	rm -rf $(_REMOTE_GNO_ROOT)
	git clone --depth 1 --branch $(GH_BRANCH) \
		https://github.com/$(GH_USER)/$(GH_REPO).git \
		$(_REMOTE_GNO_ROOT)
	touch $@

# Force a fresh clone regardless of the stamp file.
fetch-remote:
	rm -f $(_FETCH_STAMP)
	$(MAKE) $(_FETCH_STAMP) GH_USER=$(GH_USER) GH_REPO=$(GH_REPO) GH_BRANCH=$(GH_BRANCH)

clone-gno:
	@if [ -d "$(GNO_ROOT)" ]; then \
		echo "$(GNO_ROOT) already exists, skipping clone"; \
	else \
		git clone https://github.com/gnolang/gno.git $(GNO_ROOT); \
	fi

build-core-image: $(_FETCH_PREREQ)
	docker build -t $(IMAGE) -f $(CORE_GNO_ROOT)/Dockerfile --target all $(CORE_GNO_ROOT)

build-gnoland-image build-gnokey-image: build-core-image

build-gnogenesis-image: $(_FETCH_PREREQ)
	docker build -t $(GNOGENESIS_IMAGE) -f $(GNOGENESIS_GNO_ROOT)/Dockerfile --target gnocontribs $(GNOGENESIS_GNO_ROOT)

build-valsigner-image: $(_FETCH_PREREQ)
	docker build -t $(VALSIGNER_IMAGE) -f $(VALSIGNER_GNO_ROOT)/misc/val-scenarios/Dockerfile --target valsignerd $(VALSIGNER_GNO_ROOT)

build-images: build-core-image build-gnogenesis-image build-valsigner-image

# Binary targets for the local runtime (RUNTIME=local). These mirror the image
# targets one-for-one: same per-component source roots (CORE/GNOGENESIS/
# VALSIGNER_GNO_ROOT) and the same $(_FETCH_PREREQ), so `GH_USER=...
# build-binaries` builds from the fetched fork exactly like `build-images`.
# gnoland/gnokey/valsignerd live in the root module; gnogenesis is its own.
build-core-binaries: $(_FETCH_PREREQ)
	@mkdir -p $(BIN_DIR)
	cd $(CORE_GNO_ROOT) && go build -o $(BIN_DIR)/ ./gno.land/cmd/gnoland ./gno.land/cmd/gnokey

build-gnoland-binary build-gnokey-binary: build-core-binaries

build-gnogenesis-binary: $(_FETCH_PREREQ)
	@mkdir -p $(BIN_DIR)
	cd $(GNOGENESIS_GNO_ROOT)/contribs/gnogenesis && go build -o $(BIN_DIR)/gnogenesis .

build-valsigner-binary: $(_FETCH_PREREQ)
	@mkdir -p $(BIN_DIR)
	cd $(VALSIGNER_GNO_ROOT) && go build -o $(BIN_DIR)/ ./misc/val-scenarios/cmd/valsignerd

build-binaries: build-core-binaries build-gnogenesis-binary build-valsigner-binary
	@echo "[+] built gnoland, gnokey, gnogenesis, valsignerd into $(BIN_DIR)"

clean:
	rm -rf $(WORK_ROOT)

# Live block-height monitor. Node discovery differs per RUNTIME:
#   docker: scan `docker ps` for containers publishing the RPC port.
#   local:  scan each scenario's inventory.json for validator RPC URLs.
# Both feed the same `<name>\t<rpc-url>` stream into the height probe.
status:
	@while true; do \
		printf '\033[2J\033[H'; \
		date '+%H:%M:%S'; \
		echo; \
		( $(if $(filter local,$(RUNTIME)),\
			for inv in $(WORK_ROOT)/*/inventory.json; do \
				[ -f "$$inv" ] || continue; \
				jq -r '.validators[] | "\(.name)\t\(.rpc_url)"' "$$inv" 2>/dev/null; \
			done,\
			docker ps --format '{{.Names}}\t{{.Ports}}' | grep -- '->26657/tcp' \
				| awk -F'\t' '{ port=$$2; sub(/.*0\.0\.0\.0:/,"",port); sub(/->26657.*/,"",port); print $$1"\thttp://127.0.0.1:"port }') \
		) | sort | \
		while IFS=$$'\t' read -r name url; do \
			height=$$(curl -fsS "$$url/status" 2>/dev/null \
				| jq -r '.result.sync_info.latest_block_height // "offline"' 2>/dev/null \
				|| echo offline); \
			printf '%-40s  height=%s\n' "$$name" "$$height"; \
		done; \
		sleep 2; \
	done
