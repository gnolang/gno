#!/usr/bin/env bash
# Schedule a coordinated chain halt at a block height, with a minimum version.
#
# Usage:
#   ./set-halt.sh HEIGHT MIN_VERSION
#   ./set-halt.sh 0                      # cancel a scheduled halt
#
# Example:
#   ./set-halt.sh 120000 v1.3.0
#
# MIN_VERSION must be a release tag the node parses — vMAJOR.MINOR.PATCH, or
# betanet's retired chain/gnolandMAJOR.MINOR. A floor the node cannot parse
# degrades to byte equality and refuses the upgraded binary alongside the stale
# ones, leaving the chain unable to restart. See RELEASING.md and
# gno.land/cmd/gnoland/UPGRADES.md.
#
# Environment: see README.md.
set -eo pipefail

HEIGHT="${1?Usage: $0 HEIGHT MIN_VERSION (HEIGHT 0 cancels)}"
MIN_VERSION="${2-}"

case "$HEIGHT" in
0) [ -z "$MIN_VERSION" ] || { echo "error: height 0 cancels the halt; it takes no version" >&2; exit 1; } ;;
'' | *[!0-9]*) echo "error: HEIGHT must be a non-negative block height (got '$HEIGHT')" >&2; exit 1 ;;
# Gno parses the interpolated literal below as Go does: 0120000 is 40960, and
# 00 is 0, the cancel form. Only canonical decimal reaches the proposal.
0[0-9]*) echo "error: HEIGHT must not have leading zeros (got '$HEIGHT'); Gno reads it as an octal literal" >&2; exit 1 ;;
*)
  [ -n "$MIN_VERSION" ] || { echo "error: MIN_VERSION is required for a halt (height 0 is the cancel form)" >&2; exit 1; }
  # The shape gno.land/pkg/gnoland.parseReleaseVersion accepts, spelled as an
  # ERE; kept identical to misc/release/cut-release.sh's VERSION_RE, and held to
  # the Go parser by gno.land/pkg/gnoland.TestReleaseToolingMatchesTheParser.
  # Keep it on one line. The legacy alternative is betanet's two retired tags.
  VERSION_RE='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*)?$'
  if ! [[ $MIN_VERSION =~ $VERSION_RE || $MIN_VERSION =~ ^chain/gnoland(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
    echo "error: MIN_VERSION must be vMAJOR.MINOR.PATCH (got '$MIN_VERSION')." >&2
    echo "       A floor the node cannot parse degrades to byte equality and refuses" >&2
    echo "       the upgraded binary too, leaving the chain unable to restart." >&2
    exit 1
  fi
  ;;
esac

GNOKEY_NAME="${GNOKEY_NAME:?GNOKEY_NAME is required}"
CHAIN_ID="${CHAIN_ID:?CHAIN_ID is required}"
REMOTE="${REMOTE:?REMOTE is required}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/set_halt.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/sys/params"
)

func main(cur realm) {
	req := params.NewSetHaltRequest(cross(cur), ${HEIGHT}, "${MIN_VERSION}")
	dao.MustCreateProposal(cross(cur), req)
}
GOEOF

if [ "$HEIGHT" = "0" ]; then
  echo "Cancelling the scheduled halt and clearing the minimum version:"
else
  echo "Scheduling a coordinated halt:"
  echo "  Height:      ${HEIGHT}"
  echo "  Min version: ${MIN_VERSION}"
fi
echo "  Key:    ${GNOKEY_NAME}"
echo "  Chain:  ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""
echo "Proposal body:"
sed 's/^/    /' "$TMPDIR/set_halt.gno"
echo ""

gnokey maketx run \
  -gas-wanted "$GAS_WANTED" \
  -gas-fee "$GAS_FEE" \
  -broadcast \
  -chainid "$CHAIN_ID" \
  -remote "$REMOTE" \
  "$GNOKEY_NAME" \
  "$TMPDIR/set_halt.gno"

echo ""
echo "Done — proposal created. Members vote and execute it separately."
if [ "$HEIGHT" != "0" ]; then
  echo "Before it passes, every operator should confirm halt_height is unset in"
  echo "their own config.toml: a future-dated value there halts at the wrong block."
fi
