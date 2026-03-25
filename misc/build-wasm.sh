#!/usr/bin/env bash
# misc/build-wasm.sh — Build gno.wasm and root.zip locally
#
# Usage: ./misc/build-wasm.sh [output-dir]
# Default output-dir: gnovm/build/
#
# Requirements: go (with GOARCH=wasm GOOS=js support), zip
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="${1:-$REPO_ROOT/gnovm/build}"

mkdir -p "$OUTPUT_DIR"

echo "==> Building gno.wasm..."
cd "$REPO_ROOT/gnovm"
GOARCH=wasm GOOS=js go build \
  -ldflags "-X github.com/gnolang/gno/gnovm/pkg/gnoenv._GNOROOT=$REPO_ROOT/" \
  -o "$OUTPUT_DIR/gno.wasm" \
  ./cmd/gno
echo "    $(du -sh "$OUTPUT_DIR/gno.wasm" | cut -f1)  $OUTPUT_DIR/gno.wasm"

echo "==> Creating root.zip..."
cd "$REPO_ROOT"
zip -qq -i "*.toml" -i "*.gno" \
  -x "*_test.gno" -x "*_filetest.gno" \
  -r "$OUTPUT_DIR/root.zip" \
  gnovm/stdlibs gnovm/tests/stdlibs examples
echo "    $(du -sh "$OUTPUT_DIR/root.zip" | cut -f1)  $OUTPUT_DIR/root.zip"

echo "==> Done."
