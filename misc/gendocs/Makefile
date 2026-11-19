all: clean gen

# Pin to v0.2.0: v0.3.0+ requires Go >= 1.26, but CI pins GOTOOLCHAIN=local to
# the go.mod toolchain (currently 1.25.x), so @latest fails to build.
install:
	go install golang.org/x/pkgsite/cmd/pkgsite@v0.2.0

gen:
	./gendocs.sh

clean:
	rm -rf godoc

kill_zombies:
	kill -9 `lsof -t -i tcp:8080 -s TCP:LISTEN` || true
