all: clean gen

install:
	go install golang.org/x/pkgsite/cmd/pkgsite@latest

gen:
	./gendocs.sh

clean:
	rm -rf godoc

kill_zombies:
	kill -9 `lsof -t -i tcp:8080 -s TCP:LISTEN` || true
