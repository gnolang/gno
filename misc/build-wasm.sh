#!/usr/bin/env bash
# misc/build-wasm.sh — Build gno.wasm and root.zip from gnovm/cmd/gno
#
# Usage: ./misc/build-wasm.sh [--push] [output-dir]
#   --push        Publish assets to the GitHub release for the current tag
#   output-dir    Where to write gno.wasm and root.zip (default: gnovm/build/)
#
# Requirements: go (GOARCH=wasm GOOS=js), zip
# For --push: gh (GitHub CLI) authenticated with write access to gnolang/gno
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

PUSH=false
OUTPUT_DIR=""

for arg in "$@"; do
  case "$arg" in
    --push) PUSH=true ;;
    *) OUTPUT_DIR="$arg" ;;
  esac
done

OUTPUT_DIR="${OUTPUT_DIR:-$REPO_ROOT/gnovm/build}"
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

if ! $PUSH; then
  echo "==> Done. (pass --push to publish to GitHub release)"
  exit 0
fi

TAG="$(git -C "$REPO_ROOT" describe --exact-match HEAD 2>/dev/null || echo "")"
if [ -z "$TAG" ]; then
  echo "ERROR: HEAD is not on a tag. Cannot determine release name."
  exit 1
fi

ENCODED_TAG="$(python3 -c "import urllib.parse, sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$TAG")"

echo "==> Publishing to GitHub release: $TAG"
if gh release view "$TAG" --repo gnolang/gno &>/dev/null; then
  echo "    Release exists, uploading assets..."
else
  echo "    Creating release..."
  gh release create "$TAG" \
    --repo gnolang/gno \
    --title "GnoVM $TAG" \
    --notes "GnoVM WebAssembly build for tag \`$TAG\`.

Built from [gnovm/cmd/gno](https://github.com/gnolang/gno/tree/$TAG/gnovm/cmd/gno).

## Assets
- \`gno.wasm\` — GnoVM WebAssembly binary (GOOS=js GOARCH=wasm)
- \`root.zip\` — Standard libraries and examples (stdlibs + tests/stdlibs + examples)"
fi

gh release upload "$TAG" \
  --repo gnolang/gno \
  --clobber \
  "$OUTPUT_DIR/gno.wasm" \
  "$OUTPUT_DIR/root.zip"

echo ""
echo "==> Done!"
echo "    gno.wasm: https://github.com/gnolang/gno/releases/download/${ENCODED_TAG}/gno.wasm"
echo "    root.zip: https://github.com/gnolang/gno/releases/download/${ENCODED_TAG}/root.zip"
