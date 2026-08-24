#!/usr/bin/env bash
# misc/build-wasm.sh — Build gno.wasm and root.zip from gnovm/cmd/gno
#
# Requirements: go (GOARCH=wasm GOOS=js), zip
# For --push: gh (GitHub CLI) with write access to gnolang/gno
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

usage() {
  cat <<USAGE
Usage: $(basename "$0") <tag> [--push] [output-dir]

  tag         Git tag or ref to build (use HEAD for current checkout)
  --push      Publish assets to the GitHub release for the given tag
  output-dir  Where to write gno.wasm and root.zip (default: gnovm/build/)

Examples:
  $(basename "$0") chain/gnoland1.0
  $(basename "$0") chain/gnoland1.0 --push
  $(basename "$0") HEAD --push
  $(basename "$0") chain/gnoland1.0 --push ./out/
USAGE
}

if [ $# -eq 0 ]; then
  usage
  exit 0
fi

TAG=""
PUSH=false
OUTPUT_DIR=""

for arg in "$@"; do
  case "$arg" in
    --push) PUSH=true ;;
    --help|-h) usage; exit 0 ;;
    *)
      if [ -z "$TAG" ]; then
        TAG="$arg"
      else
        OUTPUT_DIR="$arg"
      fi
      ;;
  esac
done

if [ -z "$TAG" ]; then
  echo "ERROR: <tag> is required."
  echo ""
  usage
  exit 1
fi

OUTPUT_DIR="${OUTPUT_DIR:-$REPO_ROOT/gnovm/build}"
mkdir -p "$OUTPUT_DIR"

# Resolve tag: if not HEAD, checkout the tag (detached) then restore
ORIG_HEAD="$(git -C "$REPO_ROOT" symbolic-ref --short HEAD 2>/dev/null || git -C "$REPO_ROOT" rev-parse HEAD)"
NEEDS_RESTORE=false

if [ "$TAG" != "HEAD" ]; then
  echo "==> Checking out $TAG..."
  git -C "$REPO_ROOT" checkout --quiet "$TAG"
  NEEDS_RESTORE=true
fi

restore() {
  if $NEEDS_RESTORE; then
    echo "==> Restoring $ORIG_HEAD..."
    git -C "$REPO_ROOT" checkout --quiet "$ORIG_HEAD"
  fi
}
trap restore EXIT

echo "==> Building gno.wasm (tag: $TAG)..."
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

RELEASE_TAG="$TAG"
if [ "$TAG" = "HEAD" ]; then
  RELEASE_TAG="$(git -C "$REPO_ROOT" describe --exact-match HEAD 2>/dev/null || true)"
  if [ -z "$RELEASE_TAG" ]; then
    echo "ERROR: HEAD is not on a tag. Use an explicit tag name with --push."
    exit 1
  fi
fi

ENCODED_TAG="$(python3 -c "import urllib.parse, sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$RELEASE_TAG")"

echo "==> Publishing to GitHub release: $RELEASE_TAG"
if gh release view "$RELEASE_TAG" --repo gnolang/gno &>/dev/null; then
  echo "    Release exists, uploading assets..."
else
  echo "    Creating release..."
  gh release create "$RELEASE_TAG" \
    --repo gnolang/gno \
    --title "GnoVM $RELEASE_TAG" \
    --notes "GnoVM WebAssembly build for tag \`$RELEASE_TAG\`.

Built from [gnovm/cmd/gno](https://github.com/gnolang/gno/tree/$RELEASE_TAG/gnovm/cmd/gno).

## Assets
- \`gno.wasm\` — GnoVM WebAssembly binary (GOOS=js GOARCH=wasm)
- \`root.zip\` — Standard libraries and examples (stdlibs + tests/stdlibs + examples)"
fi

gh release upload "$RELEASE_TAG" \
  --repo gnolang/gno \
  --clobber \
  "$OUTPUT_DIR/gno.wasm" \
  "$OUTPUT_DIR/root.zip"

echo ""
echo "==> Done!"
echo "    gno.wasm: https://github.com/gnolang/gno/releases/download/${ENCODED_TAG}/gno.wasm"
echo "    root.zip: https://github.com/gnolang/gno/releases/download/${ENCODED_TAG}/root.zip"
