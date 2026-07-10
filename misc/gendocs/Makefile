all: clean gen

install:
	# Pinned: pkgsite v0.3.0+ requires go >= 1.26; CI's toolchain follows
	# go.mod (go-version-file). Revert to @latest once go.mod is >= 1.26.
	go install golang.org/x/pkgsite/cmd/pkgsite@v0.2.0

gen:
	./gendocs.sh

clean:
	rm -rf godoc

kill_zombies:
	kill -9 `lsof -t -i tcp:8080 -s TCP:LISTEN` || true
