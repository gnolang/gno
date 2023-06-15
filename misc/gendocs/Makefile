all: clean gen

gen:
	./gendocs.sh

clean:
	rm -rf godoc

kill_zombies:
	kill -9 `lsof -t -i tcp:6060 -s TCP:LISTEN` || true
