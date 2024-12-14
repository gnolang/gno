all: clean gen

install:
	go install .

gen:
	stdlib_diff -src $(GOROOT)/src -dst ../../gnovm/stdlibs -out ./stdlib_diff

clean:
	rm -rf stdlib_diff
