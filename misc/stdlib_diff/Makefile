all: clean gen

gen:
	go run . -src $(GOROOT)/src -dst ../../gnovm/stdlibs -out ./stdlib_diff

clean:
	rm -rf stdlib_diff
