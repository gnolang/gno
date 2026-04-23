#!/usr/bin/env bash
# compare-gas-modes.sh — run the in-memory smoketest twice against the
# current genesis, once with GasReplayMode="strict" (new VM gas meter
# applied to historical txs) and once with "source" (bypass the gas meter
# for historical txs, preserving source-chain outcomes), and report the
# difference in replay failure counts.
#
# Why
# ===
#   Most of the 2580 InsufficientFundsError failures we observe on rc6 are
#   cascades from post-mainnet storage-deposit charges that didn't exist
#   on the source chain. "source" mode is designed for exactly this: it
#   bypasses the new gas meter for historical txs so their outcomes match
#   what the source chain recorded. Before recommending "source" for a
#   production launch we need concrete numbers on how many failures it
#   actually eliminates vs "strict".
#
#   gnogenesis fork generate currently writes gas_replay_mode="" (which
#   behaves as "strict"). This script patches the generated genesis.json
#   to toggle the field, runs the smoketest on each variant, and diffs the
#   reported failure count. It doesn't modify the authoritative genesis.
#
# Env
# ===
#   OUT        misc/hf-glue/out (auto-resolved)
#   REPO       gno repo root (auto-resolved)
#   GENESIS    path to the already-built genesis (default $OUT/genesis.json)
#
# Exit
# ====
#   0  — both smoketests ran, report written
#   2  — prerequisite error (missing genesis, jq, etc.)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HERE="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO="${REPO:-$(cd "$HERE/../.." && pwd)}"
OUT="${OUT:-$HERE/out}"
GENESIS="${GENESIS:-$OUT/genesis.json}"

command -v jq >/dev/null 2>&1 || {
  echo "jq not found" >&2
  exit 2
}
[[ -f "$GENESIS" ]] || {
  echo "genesis not found at $GENESIS" >&2
  exit 2
}

REPORT="$OUT/GAS-MODES-COMPARE.md"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

echo "━━━ compare-gas-modes ━━━"
echo "  genesis  $GENESIS"
echo

run_smoketest() {
  local mode="$1" gen="$2" log="$3"
  echo "  running smoketest with gas_replay_mode=\"$mode\""
  # gnogenesis fork test exits non-zero when failures > 0 (which is the
  # common case with --skip-failing-genesis-txs absorbing noise), so we
  # tolerate that here and extract the count from the output.
  (cd "$REPO/contribs/gnogenesis" && go run . fork test --genesis "$gen") >"$log" 2>&1 || true
}

extract_failures() {
  local log="$1"
  grep -oE 'Failures:[[:space:]]+[0-9]+' "$log" | head -1 | grep -oE '[0-9]+'
}

extract_ok() {
  local log="$1"
  grep -oE 'Txs processed:[[:space:]]+[0-9]+[[:space:]]*/[[:space:]]*[0-9]+' "$log" | head -1
}

# Variant A — strict (empty string; app.go treats "" and "strict" as same)
jq '.app_state.gas_replay_mode = "strict"' "$GENESIS" >"$WORK/strict.json"
run_smoketest strict "$WORK/strict.json" "$WORK/strict.log"

# Variant B — source
jq '.app_state.gas_replay_mode = "source"' "$GENESIS" >"$WORK/source.json"
run_smoketest source "$WORK/source.json" "$WORK/source.log"

s_fail="$(extract_failures "$WORK/strict.log")"
s_tx="$(extract_ok "$WORK/strict.log")"
u_fail="$(extract_failures "$WORK/source.log")"
u_tx="$(extract_ok "$WORK/source.log")"

delta=$((${s_fail:-0} - ${u_fail:-0}))

{
  echo "# Gas replay mode comparison"
  echo ""
  echo "_Generated $(date -u +%Y-%m-%dT%H:%M:%SZ)_"
  echo ""
  echo "- **Genesis**: \`$GENESIS\`"
  echo ""
  echo "| Mode   | Txs processed                          | Failures |"
  echo "|--------|----------------------------------------|---------:|"
  echo "| strict | ${s_tx:-<unknown>}                     | ${s_fail:-?} |"
  echo "| source | ${u_tx:-<unknown>}                     | ${u_fail:-?} |"
  echo ""
  echo "**Delta (strict − source): ${delta}** failures eliminated by bypassing the new-VM gas meter for historical txs."
  echo ""
  echo "Interpretation"
  echo "--------------"
  echo ""
  echo "- \`strict\` (default) applies the current VM's gas meter to every"
  echo "  tx, including historical ones. Txs that consumed N gas on the"
  echo "  source chain may consume >N on this branch (new opcodes,"
  echo "  storage-deposit metering) and fail under the original fee cap."
  echo "- \`source\` skips gas metering for txs with \`metadata.block_height > 0\`"
  echo "  (historical) and records \`metadata.gas_used\` from the source"
  echo "  chain in the response. Result: historical txs match source"
  echo "  outcomes regardless of VM gas drift."
  echo ""
  echo "Launch posture"
  echo "--------------"
  echo ""
  if [[ "$delta" -gt 0 ]]; then
    echo "\`source\` mode eliminated ${delta} failure(s). If the eliminated"
    echo "failures are InsufficientFundsError (fee-cap shortfalls) or"
    echo "gas-related, it's preserving user interactions that would"
    echo "otherwise be lost. Trade-off: any genuinely-broken historical"
    echo "tx (VM panic unrelated to gas) is still skipped, and the new"
    echo "VM's gas meter is bypassed for history so post-fork chains"
    echo "can't rely on that metering for replay-time audits."
  else
    echo "\`source\` mode eliminated **zero** failures on this genesis."
    echo "The 2580-ish InsufficientFundsError replay failures we observe"
    echo "fire at the ante handler (\`DeductFees\`) BEFORE the gas meter"
    echo "is consulted — DeductFees checks the signer's static GasFee"
    echo "against their live balance, and a signer drained by"
    echo "post-mainnet storage-deposit charges has 0 ugnot regardless"
    echo "of which gas-replay mode the VM would use afterwards."
    echo ""
    echo "Recovering these failures cannot be done through gas-replay"
    echo "mode. The available levers are:"
    echo "  1. \`hf_topup_balance\` for each diverged signer (see"
    echo "     \`make audit-balances\`)."
    echo "  2. Patching \`DeductFees\` itself to skip historical txs"
    echo "     under a SkipFeeDeduction context key — bigger semantic"
    echo "     change, upstream-only."
    echo ""
    echo "For test13-rc the default \`strict\`/empty mode is fine; \`source\`"
    echo "adds no value on top and makes the chain harder to audit."
  fi
} >"$REPORT"

echo
echo "Report: $REPORT"
echo "strict fail=$s_fail | source fail=$u_fail | delta=$delta"
