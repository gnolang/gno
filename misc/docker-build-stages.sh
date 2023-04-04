#!/usr/bin/env bash

set -e

CURDIR=$(dirname "$0")
ROOTDIR=$(realpath --relative-to="$(pwd)" "$CURDIR")/..

set -x

docker build --target=gnoland-slim   -t ghcr.io/gnoland/gno/gnoland-slim   "$ROOTDIR"
docker build --target=gnokey-slim    -t ghcr.io/gnoland/gno/gnokey-slim    "$ROOTDIR"
docker build --target=gno-slim       -t ghcr.io/gnoland/gno/gno-slim       "$ROOTDIR"
docker build --target=gnofaucet-slim -t ghcr.io/gnoland/gno/gnofaucet-slim "$ROOTDIR"
docker build --target=gnoweb-slim    -t ghcr.io/gnoland/gno/gnoweb-slim    "$ROOTDIR"
docker build                         -t ghcr.io/gnoland/gno                "$ROOTDIR"

docker images | grep ghcr.io/gnoland
