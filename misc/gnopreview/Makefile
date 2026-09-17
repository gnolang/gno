.PHONY: test run clean install

GNOROOT ?= $(shell git rev-parse --show-toplevel)
GNODEV  ?= /tmp/gnodev
BASE    ?= origin/master
OUT     ?= _preview
PORT    ?= 8777

test:
	go test ./...

# Render the preview for the current branch, then serve it.
run: $(GNODEV)
	git -C $(GNOROOT) diff --name-only $(BASE)... > $(OUT).changed
	GNODEV=$(GNODEV) GNOROOT=$(GNOROOT) go run . render \
		-root $(GNOROOT) -changed $(OUT).changed -out $(OUT)
	@echo "serving $(OUT) on http://localhost:$(PORT)/"
	@cd $(OUT) && python3 -m http.server $(PORT)

$(GNODEV):
	cd $(GNOROOT)/contribs/gnodev && go build -o $(GNODEV) .

install:
	go install .

clean:
	rm -rf $(OUT) $(OUT).changed
