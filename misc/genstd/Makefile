run:
	cd ../../gnovm/stdlibs && go run ../../misc/genstd
	cd ../../gnovm/tests/stdlibs && go run ../../../misc/genstd -skip-init-order

test:
	go test -v .

test.update:
	go test -v . -update-golden-tests
