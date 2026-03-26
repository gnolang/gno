#!/bin/bash
set -euo pipefail

TAG="${1:?Usage: deploy.sh <tag> (e.g. chain/gnoland1.0)}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PLAYGROUND_DIR="$SCRIPT_DIR/../../../contribs/playground"

cd "$PLAYGROUND_DIR"

echo "Building playground for tag: $TAG"
make build-release TAG="$TAG"

echo "Build complete. Output in: $PLAYGROUND_DIR/app/dist/"
echo "Deploy the contents of app/dist/ to your hosting provider."
