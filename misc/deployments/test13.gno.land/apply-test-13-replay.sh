#!/usr/bin/env bash
# apply-test-13-replay.sh — phase 2 of the test-13 genesis build.
#
# Takes the base genesis produced by build-test-13-genesis.sh and:
#   1. Builds the gnokey + gnogenesis binaries from the worktree.
#   2. Verifies the txs source: for --source-txs-jsonl-file, confirms the
#      cached archive's max BlockHeight matches HALT_HEIGHT; for
#      --source-txs-rpc / --source-txs-data-dir, gnogenesis enforces
#      halt-height during fetch/read.
#   3. Builds a 2-line migration .jsonl with `gnokey maketx call` + `jq`:
#        a. MsgCall to gno.land/r/test13/rotate.Rotate, signed by the
#           genesis deployer (Rotate is gated by ChainHeight()==0, not by
#           caller identity, so any signer with funds works).
#        b. MsgCall to gno.land/r/sys/names.Enable, signed by the deployer
#           but with the caller field jq-patched to the gnoland1 GovDAO
#           T1 multisig — Enable's admin check reads that field, and we
#           don't hold the multisig key. Genesis replay runs under
#           --skip-genesis-sig-verification, so the dummy deployer
#           signature is trusted-but-ignored.
#   4. Calls `gnogenesis fork generate` to assemble the final genesis:
#        base + gnoland1 history (with GasReplayMode="source" so historical
#        txs use original-chain gas accounting, set in buildHardforkGenesis)
#        + the valoper-seed + the T1/names migration txs.
#
# After Step 4 the audit phase runs `gnogenesis fork test --verbose
# --skip-failing-genesis-txs` and categorizes any failures. 5 known
# buckets are tolerated (each documented inline in EXPECTED_PATTERNS):
#   - r/sys/validators/v2 remove (v2 store unseeded on test-13 by design)
#   - r/gnops/valopers squat guard added post-gnoland1
#   - r/gnoland/boards2/v1 owner list narrowed on master
#   - boards2 missing-board cascade (consequence of the previous)
#   - boards2 missing-member cascade (consequence of the previous two)
# Anything else is printed with full context and the audit step exits
# non-zero.
#
# Inputs (consumed; must exist or be reachable before run):
#   ./out/base-genesis.json   produced by ./build-test-13-genesis.sh — used as --source-genesis-file
#                             (carries the test-13 valset, packages, govdao setup, faucets, and a
#                              state.Balances byte-identical to gnoland1's up to the airdrop tail).
#   ./out/valoper-seed.jsonl  produced by ./build-test-13-genesis.sh — appended as a --migration-tx
#                             (one valopers.Register tx per INITIAL_VALSET entry; required by #5701/#5702).
#
# Txs source (default: multi-endpoint RPC fetch against $DEFAULT_TXS_RPC_ENDPOINTS;
# pass one of the flags below to override):
#   --source-txs-jsonl-file PATH cached amino-JSONL of gnoland.TxWithMetadata
#   --source-txs-rpc URLS        multi-endpoint RPC fetch (#5693 parallel fetcher), comma-separated
#   --source-txs-data-dir PATH   read txs from a halted gnoland data dir (offline PebbleDB reader, #5696)
#
# Output:
#   ./out/genesis.json        the final test-13 hardfork genesis
#   ./out/t1-rotation.jsonl   the 2 T1+names migration txs (kept for audit)
#   ./out/fork-test.log       full `fork test --verbose` log (kept for audit)
#
# Usage:
#   ./apply-test-13-replay.sh                                  # default: multi-RPC tx fetch
#   ./apply-test-13-replay.sh --debug                          # show every command being run
#   ./apply-test-13-replay.sh --no-install                     # reuse previously built binaries
#   ./apply-test-13-replay.sh --skip-audit                     # skip the fork-test audit step
#   ./apply-test-13-replay.sh --source-txs-jsonl-file PATH     # use a pre-fetched jsonl instead of RPC
#   ./apply-test-13-replay.sh --source-txs-data-dir PATH       # use a halted gnoland data dir instead of RPC
set -eo pipefail

# =============================================================================
# Launch parameters — review before each genesis generation.
# =============================================================================

CHAIN_ID=test-13
ORIGINAL_CHAIN_ID=gnoland1

# Source-chain halt height. Passed to gnogenesis fork generate as
# --halt-height and enforced on the fetch/read side for every txs source.
# In --source-txs-jsonl-file mode, the script also cross-checks the cached
# archive's max BlockHeight against this constant to fail fast on a stale
# cache. The resulting chain starts at InitialHeight = HALT_HEIGHT + 1.
HALT_HEIGHT=1485629

# Default RPC endpoint list for --source-txs-rpc. Comma-separated for
# #5693's multi-endpoint parallel fetcher. Used when no --source-txs-*
# flag is passed.
DEFAULT_TXS_RPC_ENDPOINTS="http://51.159.14.234:26657,http://163.172.33.181:26657,https://rpc.gnoland1.moul.p2p.team,https://rpc.gnoland1-aeddi-1.gnoland.network,https://rpc.gnoland1-gfanton-1.gnoland.network"

# =============================================================================
# Internal — everything below is glue, you shouldn't need to change it.
# =============================================================================

# ---- Flags

DEBUG=false
NO_INSTALL=false
SKIP_AUDIT=false
SOURCE_TXS_JSONL_FILE=""
SOURCE_TXS_RPC=""
SOURCE_TXS_DATA_DIR=""

# Validate that the option being parsed has an accompanying value (--key value form).
require_arg() {
  if [ "$#" -lt 2 ]; then
    echo "ERROR: $1 requires a value" >&2
    exit 1
  fi
}

while [ $# -gt 0 ]; do
  case "$1" in
  --debug)
    DEBUG=true
    shift
    ;;
  --no-install)
    NO_INSTALL=true
    shift
    ;;
  --skip-audit)
    SKIP_AUDIT=true
    shift
    ;;
  --source-txs-jsonl-file)
    require_arg "$@"
    SOURCE_TXS_JSONL_FILE="$2"
    shift 2
    ;;
  --source-txs-jsonl-file=*)
    SOURCE_TXS_JSONL_FILE="${1#*=}"
    shift
    ;;
  --source-txs-rpc)
    require_arg "$@"
    SOURCE_TXS_RPC="$2"
    shift 2
    ;;
  --source-txs-rpc=*)
    SOURCE_TXS_RPC="${1#*=}"
    shift
    ;;
  --source-txs-data-dir)
    require_arg "$@"
    SOURCE_TXS_DATA_DIR="$2"
    shift 2
    ;;
  --source-txs-data-dir=*)
    SOURCE_TXS_DATA_DIR="${1#*=}"
    shift
    ;;
  *)
    echo "Unknown argument: $1" >&2
    exit 1
    ;;
  esac
done

# Resolve txs source — default to local txs.jsonl if no flag given.
TXS_SOURCE_COUNT=0
[ -n "$SOURCE_TXS_JSONL_FILE" ] && TXS_SOURCE_COUNT=$((TXS_SOURCE_COUNT + 1))
[ -n "$SOURCE_TXS_RPC" ] && TXS_SOURCE_COUNT=$((TXS_SOURCE_COUNT + 1))
[ -n "$SOURCE_TXS_DATA_DIR" ] && TXS_SOURCE_COUNT=$((TXS_SOURCE_COUNT + 1))
if [ "$TXS_SOURCE_COUNT" -gt 1 ]; then
  echo "ERROR: --source-txs-{jsonl-file,rpc,data-dir} are mutually exclusive (pick one)." >&2
  exit 1
fi

run() {
  if [ "$DEBUG" = true ]; then
    printf "    \033[2m\$ %s\033[0m\n" "$*" >&2
  fi
  "$@"
}

# ---- Derived paths

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_GENESIS="$SCRIPT_DIR/out/base-genesis.json"
OUT_GENESIS="$SCRIPT_DIR/out/genesis.json"
OUT_MIGRATIONS="$SCRIPT_DIR/out/t1-rotation.jsonl"
VALOPER_SEED="$SCRIPT_DIR/out/valoper-seed.jsonl"
OUT_FORK_TEST_LOG="$SCRIPT_DIR/out/fork-test.log"
WORK_DIR="$SCRIPT_DIR/genesis-work"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
GNOKEY_CMD="$REPO_ROOT/gno.land/cmd/gnokey"
GNOGENESIS_CMD="$REPO_ROOT/contribs/gnogenesis"
WORK_DIR_BIN="$WORK_DIR/bin"
GNOKEY_BIN="$WORK_DIR_BIN/gnokey"
GNOGENESIS_BIN="$WORK_DIR_BIN/gnogenesis"

# r/sys/names admin: hardcoded in examples/gno.land/r/sys/names/verifier.gno
# (the gnoland1 GovDAO T1 multisig). names.Enable's admin check reads
# runtime.PreviousRealm().Address(); under --skip-genesis-sig-verification,
# the chain trusts the MsgCall.Caller field as the EOA, so jq-patching
# caller to this address makes Enable's gate pass.
NAMES_ADMIN=g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh

# Deployer key reused from phase 1's genesis-work/gnokey-home. Same mnemonic
# as gnoland1's deployer (build-test-13-genesis.sh DEPLOYER_MNEMONIC) so the
# address is reproducible.
DEPLOYER_KEY=GenesisDeployer
DEPLOYER_GNOKEY_HOME="$WORK_DIR/gnokey-home"

mkdir -p "$SCRIPT_DIR/out" "$WORK_DIR_BIN"

# Pre-flight checks: resolve the effective txs source and verify required inputs exist.

if [ ! -f "$BASE_GENESIS" ]; then
  echo "ERROR: $BASE_GENESIS not found — run ./build-test-13-genesis.sh first." >&2
  exit 1
fi

if [ ! -f "$VALOPER_SEED" ]; then
  echo "ERROR: $VALOPER_SEED not found — run ./build-test-13-genesis.sh first." >&2
  echo "       (build step 6 emits this file: one valopers.Register tx per INITIAL_VALSET entry)." >&2
  exit 1
fi

# ---- Txs source: default to multi-RPC fetch against DEFAULT_TXS_RPC_ENDPOINTS.
# The user can override by passing --source-txs-jsonl-file or
# --source-txs-data-dir; in those modes the path/dir must exist.
if [ "$TXS_SOURCE_COUNT" -eq 0 ]; then
  SOURCE_TXS_RPC="$DEFAULT_TXS_RPC_ENDPOINTS"
fi
if [ -n "$SOURCE_TXS_JSONL_FILE" ] && [ ! -f "$SOURCE_TXS_JSONL_FILE" ]; then
  echo "ERROR: --source-txs-jsonl-file points at $SOURCE_TXS_JSONL_FILE which does not exist." >&2
  exit 1
fi
if [ -n "$SOURCE_TXS_DATA_DIR" ] && [ ! -d "$SOURCE_TXS_DATA_DIR" ]; then
  echo "ERROR: --source-txs-data-dir points at $SOURCE_TXS_DATA_DIR which does not exist." >&2
  exit 1
fi

# ---- 1. Build binaries

if [ "$NO_INSTALL" = true ]; then
  printf "\n=== Step 1/4: Skipping build (--no-install) ===\n"
  for bin in "$GNOKEY_BIN" "$GNOGENESIS_BIN"; do
    if [ ! -x "$bin" ]; then
      echo "ERROR: --no-install but $bin not found. Run without --no-install first." >&2
      exit 1
    fi
  done
else
  printf "\n=== Step 1/4: Building binaries ===\n"

  printf "  gnokey...     "
  run go build -C "$GNOKEY_CMD" -o "$GNOKEY_BIN" .
  printf "ok\n"

  printf "  gnogenesis... "
  run go build -C "$GNOGENESIS_CMD" -o "$GNOGENESIS_BIN" .
  printf "ok\n"
fi

# Phase 1 leaves the deployer key in $DEPLOYER_GNOKEY_HOME; if the user
# ran with a clean genesis-work/, that home is gone — abort with a clear
# message instead of mysteriously failing inside gnokey.
if [ ! -d "$DEPLOYER_GNOKEY_HOME/data/keys.db" ]; then
  echo "ERROR: deployer keybase not found at $DEPLOYER_GNOKEY_HOME/data/keys.db" >&2
  echo "       Re-run ./build-test-13-genesis.sh to repopulate it." >&2
  exit 1
fi

# ---- 2. Verify the txs source (file modes only)
# For --source-txs-rpc and --source-txs-data-dir, gnogenesis enforces
# --halt-height on the fetch side. For --source-txs-jsonl-file, we
# verify the cached archive's max BlockHeight matches HALT_HEIGHT
# before doing any work — saves an hour of replay if the cache is stale.

printf "\n=== Step 2/4: Verifying txs source ===\n"
if [ -n "$SOURCE_TXS_JSONL_FILE" ]; then
  TXS_COUNT=$(wc -l <"$SOURCE_TXS_JSONL_FILE" | tr -d ' ')
  MAX_HEIGHT=$(awk -F'"' '
    /"block_height"/ {
      for (i=1; i<=NF; i++) {
        if ($i == "block_height") {
          h = $(i+2) + 0
          if (h > max) max = h
        }
      }
    }
    END { print max+0 }
  ' "$SOURCE_TXS_JSONL_FILE")
  printf "  mode:       jsonl-file (%s)\n" "$SOURCE_TXS_JSONL_FILE"
  printf "  txs:        %s\n" "$TXS_COUNT"
  printf "  max height: %s (HALT_HEIGHT = %s)\n" "$MAX_HEIGHT" "$HALT_HEIGHT"
  if [ "$MAX_HEIGHT" -ne "$HALT_HEIGHT" ]; then
    echo "ERROR: HALT_HEIGHT=$HALT_HEIGHT but txs.jsonl max BlockHeight=$MAX_HEIGHT." >&2
    echo "       Update the HALT_HEIGHT constant in this script or replace the cached jsonl." >&2
    exit 1
  fi
elif [ -n "$SOURCE_TXS_RPC" ]; then
  printf "  mode:       rpc (%s)\n" "$SOURCE_TXS_RPC"
  printf "  (gnogenesis fork generate enforces halt-height during fetch)\n"
else
  printf "  mode:       data-dir (%s)\n" "$SOURCE_TXS_DATA_DIR"
  printf "  (gnogenesis fork generate enforces halt-height during read)\n"
fi

# ---- 3. Build T1 rotation + names.Enable migration .jsonl
#
# Two genesis-mode MsgCall txs:
#   a. gno.land/r/test13/rotate.Rotate
#   b. gno.land/r/sys/names.Enable
#
# Both have their `caller` field jq-patched to the gnoland1 GovDAO T1
# multisig ($NAMES_ADMIN). Required for (b) because names.Enable checks
# `runtime.PreviousRealm().Address() == admin`. Used for (a) too only
# because the multisig is the one account guaranteed to have funds at
# migration-replay time (gnoland1 history funds it via several
# proposals to ~118 trillion ugnot before the migration step runs).
# Rotate's gate is `runtime.ChainHeight()==0`, not caller identity, so
# any caller with funds works — admin is convenient.
#
# Both signed with the deployer key from phase 1; the chain trusts the
# caller field at genesis under --skip-genesis-sig-verification, so the
# signatures are valid in shape but verification is bypassed — which is
# why patching caller post-sign is safe.
#
# Why we don't sign with the deployer's caller: the deployer is funded
# only for the genesis-mode txs of phase 1 (step 4 calculates the exact
# fee total and credits it). They land at zero after phase 1, so a
# migration-tx fee of 1ugnot from the deployer would fail at
# DeductFees with std.InsufficientFundsError. Admin-as-caller dodges
# this without needing an extra balance allocation.

printf "\n=== Step 3/4: Building migration .jsonl ===\n"

# Builds one signed-but-ignored MsgCall as a TxWithMetadata jsonl line.
# Args:
#   $1 = output file (jsonl line appended)
#   $2 = pkg_path
#   $3 = func name
emit_migration_msgcall() {
  local outfile="$1"
  local pkgpath="$2"
  local funcname="$3"
  local tx_json="$WORK_DIR/migration_${funcname}.tx.json"

  echo "" | "$GNOKEY_BIN" maketx call \
    --pkgpath "$pkgpath" \
    --func "$funcname" \
    --gas-wanted 100000000 \
    --gas-fee 1ugnot \
    --chainid "$CHAIN_ID" \
    --home "$DEPLOYER_GNOKEY_HOME" \
    --broadcast=false \
    --insecure-password-stdin \
    "$DEPLOYER_KEY" >"$tx_json"

  echo "" | "$GNOKEY_BIN" sign \
    --tx-path "$tx_json" \
    --chainid "$CHAIN_ID" \
    --account-number 0 \
    --account-sequence 0 \
    --home "$DEPLOYER_GNOKEY_HOME" \
    --insecure-password-stdin \
    "$DEPLOYER_KEY" >/dev/null

  # Patch caller -> admin (multisig with funds at migration replay
  # time), wrap as TxWithMetadata jsonl line.
  jq -c --arg c "$NAMES_ADMIN" \
    '.msg[0].caller = $c | {tx: ., metadata: {block_height: "0"}}' \
    "$tx_json" >>"$outfile"
}

: >"$OUT_MIGRATIONS"
emit_migration_msgcall "$OUT_MIGRATIONS" \
  "gno.land/r/test13/rotate" "Rotate"
printf "  ✓ MsgCall %s.%s  caller=%s\n" "gno.land/r/test13/rotate" "Rotate" "$NAMES_ADMIN"

emit_migration_msgcall "$OUT_MIGRATIONS" \
  "gno.land/r/sys/names" "Enable"
printf "  ✓ MsgCall %s.%s  caller=%s\n" "gno.land/r/sys/names" "Enable" "$NAMES_ADMIN"

mig_lines=$(wc -l <"$OUT_MIGRATIONS" | tr -d ' ')
printf "  -> %s (%s migration txs)\n" "$OUT_MIGRATIONS" "$mig_lines"

# ---- 4. Assemble final genesis via gnogenesis fork generate

printf "\n=== Step 4/4: Assembling final genesis (gnogenesis fork generate) ===\n"

GEN_ARGS=(
  fork generate
  --original-chain-id "$ORIGINAL_CHAIN_ID"
  --chain-id "$CHAIN_ID"
  --halt-height "$HALT_HEIGHT"
  --migration-tx "$VALOPER_SEED"
  --migration-tx "$OUT_MIGRATIONS"
  --output "$OUT_GENESIS"
)

GEN_ARGS+=(--source-genesis-file "$BASE_GENESIS")

# Txs source (exactly one of jsonl-file / rpc / data-dir)
if [ -n "$SOURCE_TXS_JSONL_FILE" ]; then
  GEN_ARGS+=(--source-txs-jsonl-file "$SOURCE_TXS_JSONL_FILE")
elif [ -n "$SOURCE_TXS_RPC" ]; then
  GEN_ARGS+=(--source-txs-rpc "$SOURCE_TXS_RPC")
else
  GEN_ARGS+=(--source-txs-data-dir "$SOURCE_TXS_DATA_DIR")
fi

run "$GNOGENESIS_BIN" "${GEN_ARGS[@]}"

SHA=$(shasum -a 256 "$OUT_GENESIS" | awk '{print $1}')
printf "\n  sha256: %s\n" "$SHA"
printf "  -> %s (%s)\n" "$OUT_GENESIS" "$(du -h "$OUT_GENESIS" | cut -f1)"
printf "  -> %s (kept for audit)\n" "$OUT_MIGRATIONS"

# ---- 5. Audit: replay the assembled genesis in-process and categorize failures

if [ "$SKIP_AUDIT" = true ]; then
  printf "\n=== Audit skipped (--skip-audit) ===\n"
  exit 0
fi

printf "\n=== Audit: replaying genesis in-process ===\n"
printf "  Running gnogenesis fork test --verbose --skip-failing-genesis-txs...\n"
printf "  (full output: %s)\n" "$OUT_FORK_TEST_LOG"

# --skip-failing-genesis-txs absorbs failures so fork test exits 0 even with
# the expected v2 valset proposal failures. We then parse the verbose log
# ourselves and decide what's expected vs unexpected. Suppress the binary's
# stdout summary — we'll print our own.
"$GNOGENESIS_BIN" fork test \
  --genesis "$OUT_GENESIS" \
  --verbose \
  --skip-failing-genesis-txs \
  --timeout 1h \
  >"$OUT_FORK_TEST_LOG" 2>&1 &
FORK_TEST_PID=$!

# Spinner: fork test takes minutes; show progress every 30s.
spinner_idx=0
while kill -0 "$FORK_TEST_PID" 2>/dev/null; do
  elapsed=$(ps -o etime= -p "$FORK_TEST_PID" 2>/dev/null | tr -d ' ' || echo "?")
  printf "\r  ... replaying (elapsed %s) " "$elapsed"
  sleep 5
done
printf "\r%-60s\r" ""

wait "$FORK_TEST_PID" || {
  echo "ERROR: gnogenesis fork test exited non-zero." >&2
  echo "Last 30 lines of log:" >&2
  tail -30 "$OUT_FORK_TEST_LOG" >&2
  exit 1
}

# Parse the verbose log for failures. fork test emits each failure as:
#   [FAIL] height=N error=...
#   Data: <ErrorType>{<struct contents>}
# (The legacy errors.FmtError{format:"<panic>"} envelope used by PR #5653's
# original bucketing is gone; failures now wrap concrete typed errors.)
# We extract one "<Type>: <key>" line per failure into FAIL_LINES_FILE
# and bucket against EXPECTED_PATTERNS. Anything that doesn't match a
# pattern is reported as UNEXPECTED and aborts the audit.

FAIL_LINES_FILE="$WORK_DIR/fork-test-failures.txt"
grep -E "^Data: " "$OUT_FORK_TEST_LOG" | sed -E '
  s/^Data: (vm\.TypeCheckError)\{[^"]*Errors:\[\]string\{"([^"]+)".*/\1: \2/
  s/^Data: (std\.InsufficientFeeError)\{.*/\1: insufficient fee/
  s/^Data: ([A-Za-z][A-Za-z0-9_.]+)\{.*/\1: (no detail)/
' >"$FAIL_LINES_FILE" 2>/dev/null || true
TOTAL_FAILS=$(wc -l <"$FAIL_LINES_FILE" | tr -d ' ')

if [ "$TOTAL_FAILS" -eq 0 ]; then
  printf "  No failed txs.\n"
  printf "\n=== Done ===\n"
  printf "  -> %s\n" "$OUT_GENESIS"
  printf "  -> %s\n" "$OUT_MIGRATIONS"
  printf "  -> %s\n" "$OUT_FORK_TEST_LOG"
  exit 0
fi

# Each spec: pattern || label || why. Triple-pipe separator avoids
# clashing with regex chars or the explanatory text.
# Patterns match against FAIL_LINES_FILE's "<Type>: <key>" lines.
# Currently empty: PR #5653's original buckets used the pre-#5702
# errors.FmtError format and don't fire under master's typed errors.
# Add buckets here as new categories are reviewed and accepted.
EXPECTED_PATTERNS=()

# Build a single OR'd regex of all expected patterns and split out
# unexpected lines. Anything still in $UNEXPECTED_FILE is novel.
JOINED_PATTERN=""
for spec in "${EXPECTED_PATTERNS[@]}"; do
  pattern="${spec%%|||*}"
  if [ -z "$JOINED_PATTERN" ]; then
    JOINED_PATTERN="$pattern"
  else
    JOINED_PATTERN="$JOINED_PATTERN|$pattern"
  fi
done

UNEXPECTED_FILE="$WORK_DIR/fork-test-unexpected.txt"
if [ -z "$JOINED_PATTERN" ]; then
  # No expected-patterns configured — every failure is unexpected.
  cp "$FAIL_LINES_FILE" "$UNEXPECTED_FILE"
else
  grep -vE "$JOINED_PATTERN" "$FAIL_LINES_FILE" >"$UNEXPECTED_FILE" || true
fi
UNEXPECTED=$(wc -l <"$UNEXPECTED_FILE" | tr -d ' ')
EXPECTED=$((TOTAL_FAILS - UNEXPECTED))

printf "  total failed txs: %s\n" "$TOTAL_FAILS"
printf "  expected:         %s\n" "$EXPECTED"
printf "  unexpected:       %s\n" "$UNEXPECTED"

if [ "${#EXPECTED_PATTERNS[@]}" -gt 0 ]; then
  printf "\n  Expected failure breakdown:\n"
  for spec in "${EXPECTED_PATTERNS[@]}"; do
    pattern="${spec%%|||*}"
    rest="${spec#*|||}"
    label="${rest%%|||*}"
    count=$(grep -cE "$pattern" "$FAIL_LINES_FILE" || true)
    printf "    %-42s %s\n" "$label" "$count"
  done

  printf "\n  Why each bucket is expected:\n"
  for spec in "${EXPECTED_PATTERNS[@]}"; do
    pattern="${spec%%|||*}"
    rest="${spec#*|||}"
    label="${rest%%|||*}"
    why="${rest#*|||}"
    count=$(grep -cE "$pattern" "$FAIL_LINES_FILE" || true)
    if [ "$count" -gt 0 ]; then
      sample=$(grep -m1 -E "$pattern" "$FAIL_LINES_FILE" || true)
      printf "    [%s]\n      %s\n      sample: %q\n" "$label" "$why" "$sample"
    fi
  done
fi

if [ "$UNEXPECTED" -gt 0 ]; then
  printf "\n  UNEXPECTED failures (%s, no matching bucket):\n" "$UNEXPECTED"
  printf "  Top-10 by frequency:\n"
  sort "$UNEXPECTED_FILE" | uniq -c | sort -rn | head -10 |
    sed 's/^/    /'
  printf "\n  Full per-failure list: %s\n" "$UNEXPECTED_FILE"
  printf "  Full fork-test log: %s\n" "$OUT_FORK_TEST_LOG"
  exit 1
fi

printf "\n=== Done ===\n"
printf "  -> %s\n" "$OUT_GENESIS"
printf "  -> %s\n" "$OUT_MIGRATIONS"
printf "  -> %s (kept for audit)\n" "$OUT_FORK_TEST_LOG"
