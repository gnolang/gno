.PHONY: fetch report all test install clean

fetch:
	./fetch.sh

report:
	go run ./cmd/gh-report

all: fetch report

test:
	go test ./...

install:
	go install ./cmd/gh-report

clean:
	rm -rf data/
