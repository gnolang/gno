#!/bin/sh

GODOC_PORT=${GODOC_PORT:-6060}
GO_MODULE=${GO_MODULE:-github.com/gnolang/gno}
GODOC_OUT=${GODOC_OUT:-godoc}
URL=http://localhost:${GODOC_PORT}/pkg/github.com/gnolang/gno/

echo "[+] Starting godoc server..."
go run \
   -modfile ../devdeps/go.mod \
   golang.org/x/tools/cmd/godoc \
     -http="localhost:${GODOC_PORT}" &
PID=$!
# Waiting for godoc server
while ! curl --fail --silent "$URL" > /dev/null 2>&1; do
    sleep 0.1
done

echo "[+] Downloading godoc pages..."
wget \
    --recursive \
    --no-verbose \
    --convert-links \
    --page-requisites \
    --adjust-extension \
    --execute=robots=off \
    --include-directories="/lib,/pkg/$GO_MODULE,/src/$GO_MODULE" \
    --exclude-directories="*" \
    --directory-prefix="${GODOC_OUT}" \
    --no-host-directories \
    "$URL?m=all"

echo "[+] Killing godoc server..."
kill -9 "$PID"

