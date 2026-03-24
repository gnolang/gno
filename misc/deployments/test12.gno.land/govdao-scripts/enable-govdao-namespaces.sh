#!/usr/bin/env bash
# Enable basic namespace support on test12 via a two-phase bootstrap.
#
# Phase 1 — Bootstrap verifier:
#   1. Deploy a minimal bootstrap verifier at the deployer's PA namespace.
#      It extends PA-namespace logic to also allow the deployer to deploy into "sys".
#   2. DAO: switch sysnames_pkgpath to the bootstrap verifier.
#
# Phase 2 — Real v2:
#   3. Deploy r/sys/names/v2 (PA + DAO-registered namespaces).
#   4. DAO: switch sysnames_pkgpath to v2.
#
# Package sources live in ./enable-govdao-namespaces-pkg/. The __DEPLOYER_ADDR__
# placeholder is patched at runtime before any transaction is sent.
#
# Usage:
#   ./enable-govdao-namespaces.sh
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

# shellcheck source=env
source "$(dirname "$0")/env"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ---- Resolve deployer address from key ----
DEPLOYER_ADDR=$(gnokey list | grep "\. ${GNOKEY_NAME} (" | grep -o 'addr: g1[a-z0-9]*' | awk '{print $2}')
if [ -z "$DEPLOYER_ADDR" ]; then
  echo "Error: could not determine address for key '${GNOKEY_NAME}'." >&2
  exit 1
fi
echo "==> Deployer: ${GNOKEY_NAME} (${DEPLOYER_ADDR})"

# ---- Sanity check: abort if already enabled ----
echo "==> Checking if namespace support is already enabled..."
if gnokey query params/vm:p:sysnames_pkgpath --remote "$REMOTE" 2>/dev/null | grep -q 'gno\.land/r/sys/names/v2'; then
  echo "Error: sysnames_pkgpath is already set to gno.land/r/sys/names/v2. Namespace support is already enabled." >&2
  exit 1
fi

# ---- Setup: copy sources and patch placeholders ----
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cp -r "$SCRIPT_DIR/enable-govdao-namespaces-pkg/." "$TMPDIR/"

find "$TMPDIR" -type f | while read -r f; do
  sed -i '' "s|__DEPLOYER_ADDR__|$DEPLOYER_ADDR|g" "$f"
done

# ---- Phase 1: Bootstrap verifier ----

echo ""
echo "==> [1/4] Deploying bootstrap verifier at r/${DEPLOYER_ADDR}/names_bootstrap..."

gnokey maketx addpkg \
  -pkgdir "$TMPDIR/bootstrap" \
  -pkgpath "gno.land/r/${DEPLOYER_ADDR}/names_bootstrap" \
  -gas-wanted "$GAS_WANTED" \
  -gas-fee "$GAS_FEE" \
  -broadcast \
  -chainid "$CHAIN_ID" \
  -remote "$REMOTE" \
  "$GNOKEY_NAME"

echo ""
echo "==> [2/4] Switching sysnames_pkgpath to bootstrap verifier..."

cat >"$TMPDIR/switch_to_bootstrap.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/sys/params"
)

func main() {
	govExec(params.NewSysParamStringPropRequest("vm", "p", "sysnames_pkgpath",
		"gno.land/r/${DEPLOYER_ADDR}/names_bootstrap"))
}

func govExec(r dao.ProposalRequest) {
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

gnokey maketx run \
  -gas-wanted "$GAS_WANTED" \
  -gas-fee "$GAS_FEE" \
  -broadcast \
  -chainid "$CHAIN_ID" \
  -remote "$REMOTE" \
  "$GNOKEY_NAME" \
  "$TMPDIR/switch_to_bootstrap.gno"

# ---- Phase 2: Real v2 ----

echo ""
echo "==> [3/4] Deploying r/sys/names/v2..."

gnokey maketx addpkg \
  -pkgdir "$TMPDIR/sys-names" \
  -pkgpath "gno.land/r/sys/names/v2" \
  -gas-wanted "$GAS_WANTED" \
  -gas-fee "$GAS_FEE" \
  -broadcast \
  -chainid "$CHAIN_ID" \
  -remote "$REMOTE" \
  "$GNOKEY_NAME"

echo ""
echo "==> [4/4] Switching sysnames_pkgpath to v2..."

cat >"$TMPDIR/switch_to_v2.gno" <<'GOEOF'
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/sys/params"
)

func main() {
	govExec(params.NewSysParamStringPropRequest("vm", "p", "sysnames_pkgpath", "gno.land/r/sys/names/v2"))
}

func govExec(r dao.ProposalRequest) {
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

gnokey maketx run \
  -gas-wanted "$GAS_WANTED" \
  -gas-fee "$GAS_FEE" \
  -broadcast \
  -chainid "$CHAIN_ID" \
  -remote "$REMOTE" \
  "$GNOKEY_NAME" \
  "$TMPDIR/switch_to_v2.gno"

echo ""
echo "Done! Namespace support is now active on test12."
echo ""
echo "  sysnames_pkgpath  →  gno.land/r/sys/names/v2"
