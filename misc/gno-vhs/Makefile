# Generate GIFs using VHS

# Current Makefile path
CPATH := $(PWD)

# Specify on which directory to run
TAPES_DIR ?= $(GNOROOT_DIR)

# Other overwritable variables
GNOROOT_DIR ?= $(abspath $(lastword $(MAKEFILE_LIST))/../../../)
BUILD_DIR ?= $(CPATH)/.build
IMAGE_NAME ?= charmbracelet/vhs:local

# Run vhs recursively on target directory
# For more information, visit https://github.com/charmbracelet/vhs/tree/main?tab=readme-ov-file#installation
gif:
	@which ttyd ffmpeg vhs 1>/dev/null || \
		(echo "ttyd, ffmpeg, and vhs need to be installed to generate GIF"; exit 1)
	find $(TAPES_DIR) -iname '*.tape' -execdir vhs {} \; -print

## Nix
gif.nix:
	nix-shell -p ffmpeg ttyd vhs --run 'make gif'

## Docker

# Run vhs within Docker
gif.docker: bins .dockerfile.build
	docker run -it --rm \
				-v "$(BUILD_DIR):/gobin" \
				-v "$(GNOROOT_DIR):$(GNOROOT_DIR)" \
				-v "$(TAPES_DIR):$(TAPES_DIR)" \
				--workdir=$(CPATH) \
				--entrypoint="bash" \
				$(IMAGE_NAME) -c "TAPES_DIR=$(TAPES_DIR) make gif"

## Extra rules to run vhs+gno in Docker
bins: gnodev gno gnokey gnoland	

# Build binaries for Linux Docker container
gnodev: | $(BUILD_DIR)
	GOOS=linux go -C $(GNOROOT_DIR)/contribs/gnodev/cmd/gnodev build -v -o $(BUILD_DIR)/gnodev .
gno: | $(BUILD_DIR)
	GOOS=linux go -C $(GNOROOT_DIR)/gnovm/cmd/gno build -v -o $(BUILD_DIR)/gno .
gnokey: | $(BUILD_DIR)
	GOOS=linux go -C $(GNOROOT_DIR)/gno.land/cmd/gnokey build -v -o $(BUILD_DIR)/gnokey .
gnoland: | $(BUILD_DIR)
	GOOS=linux go -C $(GNOROOT_DIR)/gno.land/cmd/gnoland build -v -o $(BUILD_DIR)/gnoland .

# Force rebuild binaries
.PHONY: $(BUILD_DIR)/gnodev $(BUILD_DIR)/gnokey $(BUILD_DIR)/gno $(BUILD_DIR)/gnoland
.PHONY: bins

# Create build directory
$(BUILD_DIR):
	mkdir -p $@

.dockerfile.build: Dockerfile
	docker build -t $(IMAGE_NAME) . && touch $@

## Cleanup
clean:
	rm -rf $(BUILD_DIR)

flcean: clean
	rm -f .dockerfile.build

