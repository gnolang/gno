all: clean gen

gen:
	go run . -src $(GOROOT_SAVE)/src -dst ../../gnovm/stdlibs -out ./stdlib_diff

clean:
	rm -rf stdlib_diff
