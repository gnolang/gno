#!/usr/bin/env bash
# Submit a single govDAO proposal that applies a batch of valset changes
# (add + remove + power updates) atomically through r/sys/validators/v3.
#
# Why batch: EndBlocker computes one UpdatesFrom diff per proposal, so a
# batched proposal applies as a single tm2 valset transition. Submitting
# multiple single-change proposals instead would spread the transitions
# across several blocks and expose intermediate states (e.g. temporarily
# under-quorum) to the network.
#
# Input file format
# =================
#   One change per line, three whitespace-separated fields:
#     <address> <voting_power> [pubkey]
#
#   • voting_power=0 → REMOVE the validator; pubkey field is ignored if
#     present and may be omitted.
#   • voting_power>0 → ADD a new validator; pubkey is required.
#
#   Blank lines and lines starting with '#' are ignored. To update an
#   existing validator's power in the same batch, add a remove line
#   (power=0) followed by an add line at the new power.
#
# Usage:
#   ./batch-change.sh <changes.txt>
#
# Environment (same defaults as add-validator.sh):
#   GNOKEY_NAME   - gnokey key name (default: moul)
#   CHAIN_ID      - chain ID (default: test-13)
#   REMOTE        - RPC endpoint (default: http://127.0.0.1:26657)
#   GAS_WANTED    - gas limit (default: 100000000 — batches use more gas)
#   GAS_FEE       - gas fee (default: 2000000ugnot)
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:-moul}"
CHAIN_ID="${CHAIN_ID:-test-13}"
REMOTE="${REMOTE:-http://127.0.0.1:26657}"
GAS_WANTED="${GAS_WANTED:-100000000}"
GAS_FEE="${GAS_FEE:-2000000ugnot}"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <changes.txt>"
  echo ""
  echo "changes.txt format (one per line):"
  echo "  <address> <voting_power> [pubkey]"
  echo ""
  echo "  power=0  → remove (pubkey ignored)"
  echo "  power>0  → add (pubkey required)"
  exit 1
fi

CHANGES_FILE="$1"
[[ -f "$CHANGES_FILE" ]] || {
  echo "error: changes file not found: $CHANGES_FILE" >&2
  exit 1
}

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

# Render each non-comment, non-blank line into a Go literal inside the
# validator slice. Power=0 entries omit PubKey; power>0 entries require it.
entries=""
preview=""
lineno=0
while IFS= read -r raw || [[ -n "$raw" ]]; do
  lineno=$((lineno + 1))
  line="$(printf '%s' "$raw" | sed 's/[[:space:]]\+$//')"
  [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue

  # shellcheck disable=SC2206
  fields=($line)
  addr="${fields[0]:-}"
  power="${fields[1]:-}"
  pub="${fields[2]:-}"

  if [[ -z "$addr" || -z "$power" ]]; then
    echo "error: line $lineno: expected '<address> <power> [pubkey]', got '$raw'" >&2
    exit 1
  fi
  if ! [[ "$power" =~ ^[0-9]+$ ]]; then
    echo "error: line $lineno: power must be a non-negative integer (got '$power')" >&2
    exit 1
  fi

  # Sanity-check addresses and pubkeys so malformed entries can't inject
  # into the generated Go — only allow the bech32 character set.
  if ! [[ "$addr" =~ ^g1[0-9a-z]+$ ]]; then
    echo "error: line $lineno: address '$addr' is not a valid g1 bech32" >&2
    exit 1
  fi

  if [[ "$power" == "0" ]]; then
    entries+="$(printf '\n\t\t\t{Address: address("%s"), VotingPower: 0},' "$addr")"
    preview+=$(printf '\n  - remove %s' "$addr")
  else
    if [[ -z "$pub" ]]; then
      echo "error: line $lineno: add requires a pubkey (power=$power for $addr has no pubkey)" >&2
      exit 1
    fi
    if ! [[ "$pub" =~ ^gpub1[0-9a-z]+$ ]]; then
      echo "error: line $lineno: pubkey '$pub' is not a valid gpub1 bech32" >&2
      exit 1
    fi
    entries+="$(printf '\n\t\t\t{Address: address("%s"), PubKey: "%s", VotingPower: %s},' "$addr" "$pub" "$power")"
    preview+=$(printf '\n  - add    %s (power %s)' "$addr" "$power")
  fi
done <"$CHANGES_FILE"

if [[ -z "$entries" ]]; then
  echo "error: no changes found in $CHANGES_FILE (blank or comments only)" >&2
  exit 1
fi

cat >"$TMPDIR/batch_change.gno" <<GOEOF
package main

import (
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	valr "gno.land/r/sys/validators/v3"
)

func main() {
	executor := valr.NewValsetChangeExecutor(func() []validators.Validator {
		return []validators.Validator{${entries}
		}
	})

	r := dao.NewProposalRequest(
		"Batch valset change (${CHANGES_FILE})",
		"Apply a batch of add/remove/update changes atomically via v3.",
		executor,
	)

	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

echo "Submitting batch valset change from $CHANGES_FILE:${preview}"
echo ""
echo "  Key: ${GNOKEY_NAME}"
echo "  Chain: ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey maketx run \
  -gas-wanted "$GAS_WANTED" \
  -gas-fee "$GAS_FEE" \
  -broadcast \
  -chainid "$CHAIN_ID" \
  -remote "$REMOTE" \
  "$GNOKEY_NAME" \
  "$TMPDIR/batch_change.gno"
