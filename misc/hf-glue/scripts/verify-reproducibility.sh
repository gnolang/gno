#!/usr/bin/env bash
# verify-reproducibility.sh — assert that `make genesis` is deterministic.
#
# Two independent builders with the same code commit + same inputs must
# produce byte-identical genesis.json. This is the single hardest constraint
# for a shared-genesis hardfork: each validator must be able to rebuild from
# the published recipe and confirm the SHA256 matches what T1 attested to.
# Any nondeterminism (unsorted map iteration, timestamps, randomised keys)
# makes the whole attestation chain meaningless.
#
# This script is the local equivalent of running the recipe on two different
# machines: it stages a fresh OUT dir, builds the genesis, stages a second
# OUT dir with the same cached inputs, rebuilds, and compares.
#
# Approach
# ========
#   1. Build genesis to $OUT_A (fresh OUT dir). Cached txs.jsonl is reused
#      via $SHARED_TXS so we're not testing tx-archive's determinism (covered
#      separately by verify-txs-jsonl).
#   2. Build again to $OUT_B with identical inputs.
#   3. Compare SHA256(genesis.json) A vs B. If mismatch, print a compact
#      diff showing the first divergence hint.
#
# Determinism caveat
# ==================
#   `make genesis` is deterministic WHEN THE OUTPUT DIR IS FRESH. If you
#   run `make genesis` twice in the same $OUT without cleaning, residual
#   files from the first run (e.g. stale new_valset.json, cached
#   intermediate files that the second run would regenerate slightly
#   differently, or an ephemeral signing state) can leak into the second
#   build's SHA. Reproducibility reports in the wild that show "same
#   inputs, different SHA" almost always trace back to stale $OUT state,
#   not to real nondeterminism in the build pipeline.
#
#   This script wipes $OUT_A and $OUT_B before each build specifically to
#   rule that out. If you're debugging a cross-session SHA drift manually,
#   always `rm -rf $OUT` between attempts before blaming the tooling.
#
# Env
# ===
#   OUT_A, OUT_B   — isolated target dirs (default $OUT/reproduce-{A,B})
#   SHARED_TXS     — pre-fetched txs.jsonl to seed both builds (default
#                    $OUT/source/txs.jsonl). If unset and the file is
#                    missing, each build fetches fresh, which is slower
#                    AND introduces a per-run tx-archive step that isn't
#                    being tested here.
#   KEEP_ARTIFACTS — set to 1 to preserve $OUT_A/$OUT_B for inspection on
#                    failure (default: removed on success, kept on fail).
#   All `make genesis` env vars (HALT_HEIGHT, VALIDATOR_ADDR, VALIDATOR_PUBKEY,
#   CHAIN_ID, NEW_T1_ADDR, VALIDATOR_LIST, etc.) are forwarded to both builds.
#
# Exit status
# ===========
#   0  — identical SHA256 across both builds
#   1  — divergence detected
#   2  — prerequisite error
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HERE="$(cd "$SCRIPT_DIR/.." && pwd)"
OUT="${OUT:-$HERE/out}"

OUT_A="${OUT_A:-$OUT/reproduce-A}"
OUT_B="${OUT_B:-$OUT/reproduce-B}"
SHARED_TXS="${SHARED_TXS:-$OUT/source/txs.jsonl}"

: "${HALT_HEIGHT:?HALT_HEIGHT is required}"

command -v shasum >/dev/null 2>&1 || {
  echo "shasum not found on PATH" >&2
  exit 2
}

# On divergence we want both $OUT_A and $OUT_B preserved so the operator can
# diff them; on success they're noisy and should be cleaned up.
cleanup() {
  if [[ "${KEEP_ARTIFACTS:-0}" != 1 && "${reproduced:-0}" == 1 ]]; then
    rm -rf "$OUT_A" "$OUT_B"
  fi
}
trap cleanup EXIT

reproduced=0

build_one() {
  local out_dir="$1" label="$2"
  echo
  echo "━━━ build $label → $out_dir ━━━"
  rm -rf "$out_dir"
  mkdir -p "$out_dir/source"

  if [[ -f "$SHARED_TXS" ]]; then
    cp "$SHARED_TXS" "$out_dir/source/txs.jsonl"
    echo "  seeded cached txs.jsonl from $SHARED_TXS ($(wc -l <"$SHARED_TXS" | tr -d ' ') txs)"
  else
    echo "  WARNING: no cached txs.jsonl at $SHARED_TXS — each build will fetch (slow)"
  fi

  # OUT must be passed as a Make argument (not env) because the Makefile
  # uses `OUT := ...` which is fixed at parse time; command-line args
  # override but env does not.
  make -C "$HERE" OUT="$out_dir" genesis >"$out_dir/build.log" 2>&1 || {
    echo "  ✗ build FAILED — see $out_dir/build.log"
    exit 1
  }
  echo "  ✓ built $(du -h "$out_dir/genesis.json" | awk '{print $1}')"
}

build_one "$OUT_A" A
build_one "$OUT_B" B

sha_a="$(shasum -a 256 "$OUT_A/genesis.json" | cut -d' ' -f1)"
sha_b="$(shasum -a 256 "$OUT_B/genesis.json" | cut -d' ' -f1)"

echo
echo "━━━ compare ━━━"
echo "  A: $sha_a"
echo "  B: $sha_b"

if [[ "$sha_a" == "$sha_b" ]]; then
  echo "  ✓ reproducible"
  reproduced=1
  exit 0
fi

echo "  ✗ NON-DETERMINISTIC — same inputs, different SHA256"
echo
echo "━━━ first diff hint ━━━"
# Normalize JSON (sort keys) then byte-diff so field-order drift doesn't
# bury the real divergence.
if command -v jq >/dev/null 2>&1; then
  a_norm="$OUT_A/genesis.sorted"
  b_norm="$OUT_B/genesis.sorted"
  jq -S . "$OUT_A/genesis.json" >"$a_norm"
  jq -S . "$OUT_B/genesis.json" >"$b_norm"
  diff -u "$a_norm" "$b_norm" | head -n 40 || true
else
  echo "  (install jq to see a structured diff — raw files differ in size:)"
  wc -c "$OUT_A/genesis.json" "$OUT_B/genesis.json"
fi
echo
echo "  artifacts preserved at $OUT_A and $OUT_B for inspection"
exit 1
