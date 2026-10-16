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

# Regex bracket ranges follow the locale's collation: under en_US.UTF-8 on
# glibc, [A-Za-z] also matches letters such as é, which the node's version
# parser refuses. In the C locale the shape checks match the parser.
export LC_ALL=C

HEIGHT="${1?Usage: $0 HEIGHT MIN_VERSION (HEIGHT 0 cancels)}"
MIN_VERSION="${2-}"

case "$HEIGHT" in
0) [ -z "$MIN_VERSION" ] || {
  echo "error: height 0 cancels the halt; it takes no version" >&2
  exit 1
} ;;
'' | *[!0-9]*)
  echo "error: HEIGHT must be a non-negative block height (got '$HEIGHT')" >&2
  exit 1
  ;;
# Gno parses the interpolated literal below as Go does: 0120000 is 40960, and
# 00 is 0, the cancel form. Only canonical decimal reaches the proposal.
0[0-9]*)
  echo "error: HEIGHT must not have leading zeros (got '$HEIGHT'); Gno reads it as an octal literal" >&2
  exit 1
  ;;
*)
  [ -n "$MIN_VERSION" ] || {
    echo "error: MIN_VERSION is required for a halt (height 0 is the cancel form)" >&2
    exit 1
  }
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

# Refuse a floor the running binary already meets: it gates nothing after the
# halt and refuses every restart of that binary before it (see "Restart gates"
# in gno.land/cmd/gnoland/UPGRADES.md). The node at REMOTE stands in for the
# network.
#
# release_version, cmp_* and meets_min_version mirror parseReleaseVersion and
# meetsMinVersion in gno.land/pkg/gnoland/node_params.go, and are held to them
# by TestSetHaltRefusesAFloorTheRunningBinaryMeets.

# release_version V prints V the way the node reads it, vMAJOR.MINOR.PATCH with
# any pre-release, or fails if the node does not parse V.
release_version() {
  local v="$1"
  case "$v" in
  chain/gnoland*.*)
    # Betanet's chain/gnolandMAJOR.MINOR, widened the way the node widens it:
    # everything after the first dot is the minor, then ".0" is appended.
    v="${v#chain/gnoland}"
    v="v${v%%.*}.${v#*.}.0"
    ;;
  chain/gnoland*) return 1 ;;
  esac
  case "$v" in
  *+*)
    # Build metadata takes no part in ordering, but it must be well formed.
    [[ ${v#*+} =~ ^[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*$ ]] || return 1
    v="${v%%+*}"
    ;;
  esac
  [[ $v =~ $VERSION_RE ]] || return 1
  printf '%s\n' "$v"
}

# cmp_bytes A B prints -1, 0 or 1 for A and B in byte order: the script runs in
# the C locale, where [[ < ]] compares bytes.
cmp_bytes() {
  if [[ $1 == "$2" ]]; then
    echo 0
  elif [[ $1 < $2 ]]; then
    echo -1
  else
    echo 1
  fi
}

# cmp_numeric A B prints -1, 0 or 1 for two decimal strings without leading
# zeros, shorter first, so components wider than 64 bits order correctly.
cmp_numeric() {
  if [ "${#1}" -lt "${#2}" ]; then
    echo -1
  elif [ "${#1}" -gt "${#2}" ]; then
    echo 1
  else
    cmp_bytes "$1" "$2"
  fi
}

# cmp_prerelease A B prints -1, 0 or 1 for two pre-releases, identifier by
# identifier: numeric ones numerically and below alphanumeric ones, which
# compare in byte order; when one list is a prefix of the other, the longer
# one ranks higher.
cmp_prerelease() {
  local -a xs ys
  local i r
  IFS=. read -r -a xs <<<"$1"
  IFS=. read -r -a ys <<<"$2"
  for ((i = 0; i < ${#xs[@]} && i < ${#ys[@]}; i++)); do
    if [[ ${xs[i]} =~ ^[0-9]+$ && ${ys[i]} =~ ^[0-9]+$ ]]; then
      r=$(cmp_numeric "${xs[i]}" "${ys[i]}")
    elif [[ ${xs[i]} =~ ^[0-9]+$ ]]; then
      r=-1
    elif [[ ${ys[i]} =~ ^[0-9]+$ ]]; then
      r=1
    else
      r=$(cmp_bytes "${xs[i]}" "${ys[i]}")
    fi
    if [ "$r" != 0 ]; then
      echo "$r"
      return
    fi
  done
  cmp_numeric "${#xs[@]}" "${#ys[@]}"
}

# cmp_release A B prints -1, 0 or 1 for two versions as release_version prints
# them. A pre-release ranks below the release it leads to.
cmp_release() {
  local a="${1#v}" b="${2#v}" pa="" pb="" r
  [ "$a" = "${a%%-*}" ] || pa="${a#*-}"
  [ "$b" = "${b%%-*}" ] || pb="${b#*-}"
  # MAJOR.MINOR.PATCH is three numeric identifiers, ordered as a pre-release's.
  r=$(cmp_prerelease "${a%%-*}" "${b%%-*}")
  if [ "$r" != 0 ]; then
    echo "$r"
  elif [ -z "$pa" ] && [ -z "$pb" ]; then
    echo 0
  elif [ -z "$pa" ]; then
    echo 1
  elif [ -z "$pb" ]; then
    echo -1
  else
    cmp_prerelease "$pa" "$pb"
  fi
}

# meets_min_version BINARY FLOOR succeeds when the node would let BINARY through
# FLOOR. A version the node does not parse meets only an identical floor.
meets_min_version() {
  local b f
  if b=$(release_version "$1") && f=$(release_version "$2"); then
    [ "$(cmp_release "$b" "$f")" != -1 ]
  else
    [ "$1" = "$2" ]
  fi
}

# http_get URL prints the body at URL. gnokey dials REMOTE directly, so the
# proxy variables curl and wget honour are cleared: through a proxy the check
# could fail, or read another node than the one gnokey broadcasts to.
http_get() (
  unset http_proxy HTTP_PROXY https_proxy HTTPS_PROXY all_proxy ALL_PROXY
  if command -v curl >/dev/null 2>&1; then
    curl -fsS --max-time 10 "$1"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O - --timeout=10 --tries=1 "$1"
  else
    echo "error: reading the version REMOTE runs needs curl or wget" >&2
    return 1
  fi
)

if [ "$HEIGHT" != "0" ]; then
  case "$REMOTE" in
  http://* | https://*) STATUS_URL="${REMOTE%/}/status" ;;
  tcp://*) STATUS_URL="http://${REMOTE#tcp://}/status" ;;
  *://*)
    echo "error: cannot read /status over ${REMOTE%%://*}:// (REMOTE=$REMOTE)" >&2
    exit 1
    ;;
  *) STATUS_URL="http://${REMOTE}/status" ;;
  esac
  if ! status=$(http_get "$STATUS_URL"); then
    echo "error: cannot read the version REMOTE runs from $STATUS_URL" >&2
    exit 1
  fi
  RUNNING=""
  build_version_re='"build_version": *"([^"]*)"'
  if [[ $status =~ $build_version_re ]]; then
    RUNNING="${BASH_REMATCH[1]}"
  fi
  if [ -z "$RUNNING" ]; then
    echo "error: $STATUS_URL reports no build_version" >&2
    exit 1
  fi
  if meets_min_version "$RUNNING" "$MIN_VERSION"; then
    echo "error: $REMOTE runs $RUNNING, which already meets $MIN_VERSION." >&2
    echo "       Name the version operators switch to: a floor the running binary" >&2
    echo "       meets lets it resume after the halt, and refuses every restart of" >&2
    echo "       it before, which can stall the chain short of the halt." >&2
    exit 1
  fi
  if release_version "$RUNNING" >/dev/null; then
    echo "$REMOTE runs $RUNNING, below the floor $MIN_VERSION. This is one node:"
    echo "if the network runs mixed versions, read the others' /status too."
    echo ""
  else
    echo "warning: $REMOTE runs '$RUNNING', not a release version, so it meets no floor;" >&2
    echo "         read the other nodes' /status before relying on this one." >&2
  fi
fi

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
