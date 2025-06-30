all: clean gen

GOROOT_SAVE ?= $(shell go env GOROOT)

gen:
	go run . -src $(GOROOT_SAVE)/src -dst ../../gnovm/stdlibs -out ./stdlib_diff

clean:
	rm -rf stdlib_diff
