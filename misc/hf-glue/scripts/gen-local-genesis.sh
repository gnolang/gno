#!/usr/bin/env bash
# Rebuild the gnoland1 base genesis locally via
# misc/deployments/gnoland1/gen-genesis.sh, then stage it at
# out/source-genesis.json for the hardfork tool to consume as a file source.
#
# Inputs (env):
#   OUT    output directory (absolute)
#   REPO   repo root (absolute)
set -euo pipefail

: "${OUT:?OUT is required}"
: "${REPO:?REPO is required}"

GENESIS_SRC="$REPO/misc/deployments/gnoland1/genesis.json"
OUT_FILE="$OUT/source-genesis.json"

echo "── rebuild gnoland1 genesis locally ────────────────────────"
echo "  this runs misc/deployments/gnoland1/gen-genesis.sh (takes a few minutes)"
echo "  output will be staged at: $OUT_FILE"
echo ""

mkdir -p "$OUT"

# Reuse a pre-existing build if present to shave time on reruns.
EXTRA=()
if [[ -d "$REPO/misc/deployments/gnoland1/genesis-work/bin" ]]; then
  EXTRA+=(--no-install)
fi

( cd "$REPO/misc/deployments/gnoland1" && ./gen-genesis.sh "${EXTRA[@]}" )

if [[ ! -f "$GENESIS_SRC" ]]; then
  echo "ERROR: expected $GENESIS_SRC after gen-genesis.sh but it does not exist" >&2
  exit 1
fi

cp "$GENESIS_SRC" "$OUT_FILE"
echo ""
echo "done — source genesis at $OUT_FILE"
echo ""
echo "Next:"
echo "  SOURCE=$OUT_FILE make fetch init up"
