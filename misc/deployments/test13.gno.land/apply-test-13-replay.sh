#!/usr/bin/env bash
# apply-test-13-replay.sh — phase 2 of the test-13 genesis build.
#
# Takes the base genesis produced by build-test-13-genesis.sh and:
#   1. Verifies the cached gnoland1 historical-tx archive matches HALT_HEIGHT.
#   2. Builds a 2-line migration .jsonl with `gnokey maketx call` + `jq`:
#        a. MsgCall to gno.land/r/test13/rotate.Rotate, signed by the
#           genesis deployer (Rotate is gated by ChainHeight()==0, not by
#           caller identity, so any signer with funds works).
#        b. MsgCall to gno.land/r/sys/names.Enable, signed by the deployer
#           but with the caller field jq-patched to the gnoland1 GovDAO
#           T1 multisig — Enable's admin check reads that field, and we
#           don't hold the multisig key. Genesis replay runs under
#           --skip-genesis-sig-verification, so the dummy deployer
#           signature is trusted-but-ignored.
#   3. Calls `gnogenesis fork generate` to assemble the final genesis:
#        base + gnoland1 history (with GasReplayMode="source" baked in by
#        default at line 304-306 of fork/generate.go) + the migration txs.
#   4. Runs `gnogenesis fork test --verbose --skip-failing-genesis-txs`
#        against the result and categorizes any failures:
#          - Expected: r/sys/validators/v2 proposal-execute failures. The
#            test-13 bootstrap (govdao_prop1_test13.gno) skips the v2
#            valset seed, so gnoland1 historical txs that propose
#            add/remove against v2 find an empty store. Cosmetic — master's
#            EndBlocker reads valset from v3 + GenesisDoc.Validators, not
#            v2 events, so consensus is unaffected.
#          - Unexpected: anything else — printed with full context, audit
#            step exits non-zero.
#
# Pre-requisites:
#   ./out/base-genesis.json   produced by ./build-test-13-genesis.sh
#   ./txs.jsonl               pre-fetched gnoland1 historical txs (one
#                             gnoland.TxWithMetadata per line, amino-JSON;
#                             not downloaded by this script — see ERROR
#                             message at runtime for the recipe)
#
# Output:
#   ./out/genesis.json        the final test-13 hardfork genesis
#   ./out/t1-rotation.jsonl   the 2 migration txs (kept for audit)
#   ./out/fork-test.log       full `fork test --verbose` log (kept for audit)
#
# Usage:
#   ./apply-test-13-replay.sh              # full assemble + audit
#   ./apply-test-13-replay.sh --debug      # show every command being run
#   ./apply-test-13-replay.sh --no-install # reuse previously built binaries
#   ./apply-test-13-replay.sh --skip-audit # skip the fork-test audit step
set -eo pipefail

# =============================================================================
# Launch parameters — review before each genesis generation.
# =============================================================================

CHAIN_ID=test-13
ORIGINAL_CHAIN_ID=gnoland1

# Highest BlockHeight present in the cached txs.jsonl. The script
# verifies this against the actual jsonl content and aborts on
# mismatch — keeps the constant honest. The resulting chain starts
# at InitialHeight = HALT_HEIGHT + 1.
HALT_HEIGHT=1008282

# =============================================================================
# Internal — everything below is glue, you shouldn't need to change it.
# =============================================================================

# ---- Flags

DEBUG=false
NO_INSTALL=false
SKIP_AUDIT=false
for arg in "$@"; do
  case "$arg" in
  --debug) DEBUG=true ;;
  --no-install) NO_INSTALL=true ;;
  --skip-audit) SKIP_AUDIT=true ;;
  *)
    echo "Unknown argument: $arg"
    exit 1
    ;;
  esac
done

run() {
  if [ "$DEBUG" = true ]; then
    printf "    \033[2m\$ %s\033[0m\n" "$*" >&2
  fi
  "$@"
}

# ---- Derived paths

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_GENESIS="$SCRIPT_DIR/out/base-genesis.json"
TXS_JSONL="$SCRIPT_DIR/txs.jsonl"
OUT_GENESIS="$SCRIPT_DIR/out/genesis.json"
OUT_MIGRATIONS="$SCRIPT_DIR/out/t1-rotation.jsonl"
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

# Pre-flight checks before we touch anything.
if [ ! -f "$BASE_GENESIS" ]; then
  echo "ERROR: $BASE_GENESIS not found — run ./build-test-13-genesis.sh first."
  exit 1
fi
if [ ! -f "$TXS_JSONL" ]; then
  cat >&2 <<EOF
ERROR: $TXS_JSONL not found.

The script expects a pre-fetched gnoland1 historical-tx archive at
that path. The file is one gnoland.TxWithMetadata per line in amino-
JSON form. It is NOT downloaded by this script — fetching ~1M blocks
from rpc.gno.land sequentially takes hours, so we treat it as an
external input.

To produce it:
  go run -C contribs/gnogenesis . fork generate \\
    --source https://rpc.gno.land \\
    --halt-height $HALT_HEIGHT \\
    --chain-id $CHAIN_ID \\
    --txs-output $TXS_JSONL \\
    --output /tmp/throwaway.json

Or, if you already have a cached file from an earlier run, copy it
to $TXS_JSONL.
EOF
  exit 1
fi

# ---- 1. Build binaries

if [ "$NO_INSTALL" = true ]; then
  printf "\n=== Step 1/4: Skipping build (--no-install) ===\n"
  for bin in "$GNOKEY_BIN" "$GNOGENESIS_BIN"; do
    if [ ! -x "$bin" ]; then
      echo "ERROR: --no-install but $bin not found. Run without --no-install first."
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
  echo "ERROR: deployer keybase not found at $DEPLOYER_GNOKEY_HOME/data/keys.db"
  echo "       Re-run ./build-test-13-genesis.sh to repopulate it."
  exit 1
fi

# ---- 2. Verify the cached txs.jsonl matches the HALT_HEIGHT we declared
# The constant at the top has to match the highest BlockHeight in the
# jsonl, otherwise InitialHeight (HALT+1) lands before the latest tx.

printf "\n=== Step 2/4: Verifying cached txs.jsonl ===\n"
TXS_COUNT=$(wc -l <"$TXS_JSONL" | tr -d ' ')
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
' "$TXS_JSONL")
printf "  txs:        %s\n" "$TXS_COUNT"
printf "  max height: %s (HALT_HEIGHT = %s)\n" "$MAX_HEIGHT" "$HALT_HEIGHT"
if [ "$MAX_HEIGHT" -ne "$HALT_HEIGHT" ]; then
  echo "ERROR: HALT_HEIGHT=$HALT_HEIGHT but txs.jsonl max BlockHeight=$MAX_HEIGHT."
  echo "       Update the HALT_HEIGHT constant in this script or replace txs.jsonl."
  exit 1
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

# dirSource expects either dir/config/genesis.json or dir/genesis.json,
# plus dir/txs.jsonl. Symlinks instead of copies — base is 186MB.
SOURCE_DIR="$WORK_DIR/source"
rm -rf "$SOURCE_DIR"
mkdir -p "$SOURCE_DIR/config"
ln -sf "$BASE_GENESIS" "$SOURCE_DIR/config/genesis.json"
ln -sf "$TXS_JSONL" "$SOURCE_DIR/txs.jsonl"

run "$GNOGENESIS_BIN" fork generate \
  --source "$SOURCE_DIR" \
  --original-chain-id "$ORIGINAL_CHAIN_ID" \
  --chain-id "$CHAIN_ID" \
  --halt-height "$HALT_HEIGHT" \
  --migration-tx "$OUT_MIGRATIONS" \
  --output "$OUT_GENESIS"

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
  echo "ERROR: gnogenesis fork test exited non-zero."
  echo "Last 30 lines of log:"
  tail -30 "$OUT_FORK_TEST_LOG"
  exit 1
}

# Parse the verbose log for panic messages. fork test prints each
# failure across multiple lines:
#   [FAIL] height=N error=...
#   Data: errors.FmtError{format:"<panic message>", args:[]interface {}(nil)}
# We extract the panic text from the Data: line (more stable than the
# truncated [FAIL] line) and bucket by known regex. Every bucket maps
# to a deliberate architectural choice in this hardfork; failures not
# matching any bucket are unknown and abort.

PANIC_LINES_FILE="$WORK_DIR/fork-test-panics.txt"
grep -E "^Data: errors\.FmtError\{format:" "$OUT_FORK_TEST_LOG" |
  sed -E 's/^Data: errors\.FmtError\{format:"([^"]*)".*/\1/' >"$PANIC_LINES_FILE" 2>/dev/null || true
TOTAL_FAILS=$(wc -l <"$PANIC_LINES_FILE" | tr -d ' ')

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
EXPECTED_PATTERNS=(
  "^validator doesn't exist\$|||r/sys/validators/v2 remove|||gnoland1's bootstrap seeded v2 with 7 launch validators via govdao_prop1; test-13 skips that seed because master's EndBlocker reads valset from v3 + GenesisDoc.Validators (not v2 events). Historical proposals to remove v2 validators find an empty store and panic at the executor — no consensus impact."
  "^post-genesis: caller must equal operator address\$|||r/gnops/valopers squat guard added post-gnoland1|||master's valopers added ErrOperatorSquatGuard (commit c307ad175 'VALOPLAN2', after gnoland1 launched). Gnoland1's deployed Register had no such guard, so historical Register txs with caller != operator-addr succeeded there; on master they panic. Affected operator profiles aren't created on test-13."
  "^unauthorized, user g.+ doesn't have the required permission\$|||r/gnoland/boards2/v1 unauthorized — master narrowed owners|||gnoland1's deployed boards2 initialized gPerms with TWO owners {g16jpf…, GovDAO multisig}; master's boards2 narrowed that to {GovDAO multisig} only. g16jpf… no longer has the owner role, so their CreateBoard / realm-level InviteMember calls panic with unauthorized."
  "^board does not exist with ID: [0-9]+\$|||r/gnoland/boards2/v1 missing-board cascade|||cascade of the previous bucket: with the two CreateBoard txs failing (g16jpf no longer owner), the boards they would have created don't exist. Subsequent InviteMember/RemoveMember calls targeting those board IDs panic at mustGetBoard."
  "^member not found\$|||r/gnoland/boards2/v1 missing-member cascade|||cascade of the previous two buckets: with InviteMember failing (board missing), the invited members were never added to the board. Subsequent RemoveMember calls targeting those addresses panic at removeMember."
)

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
grep -vE "$JOINED_PATTERN" "$PANIC_LINES_FILE" >"$UNEXPECTED_FILE" || true
UNEXPECTED=$(wc -l <"$UNEXPECTED_FILE" | tr -d ' ')
EXPECTED=$((TOTAL_FAILS - UNEXPECTED))

printf "  total failed txs: %s (out of %s attempted)\n" "$TOTAL_FAILS" "$TXS_COUNT"
printf "  expected:         %s\n" "$EXPECTED"
printf "  unexpected:       %s\n" "$UNEXPECTED"

printf "\n  Expected failure breakdown:\n"
for spec in "${EXPECTED_PATTERNS[@]}"; do
  pattern="${spec%%|||*}"
  rest="${spec#*|||}"
  label="${rest%%|||*}"
  count=$(grep -cE "$pattern" "$PANIC_LINES_FILE" || true)
  printf "    %-42s %s\n" "$label" "$count"
done

printf "\n  Why each bucket is expected:\n"
for spec in "${EXPECTED_PATTERNS[@]}"; do
  pattern="${spec%%|||*}"
  rest="${spec#*|||}"
  label="${rest%%|||*}"
  why="${rest#*|||}"
  count=$(grep -cE "$pattern" "$PANIC_LINES_FILE" || true)
  if [ "$count" -gt 0 ]; then
    sample=$(grep -m1 -E "$pattern" "$PANIC_LINES_FILE" || true)
    printf "    [%s]\n      %s\n      sample panic: %q\n" "$label" "$why" "$sample"
  fi
done

if [ "$UNEXPECTED" -gt 0 ]; then
  printf "\n  UNEXPECTED failures (%s, no matching bucket):\n" "$UNEXPECTED"
  sed 's/^/    /' "$UNEXPECTED_FILE"
  printf "\n  Review %s for full context before continuing to the\n" "$OUT_FORK_TEST_LOG"
  printf "  5-node cluster test.\n"
  exit 1
fi

printf "\n=== Done ===\n"
printf "  -> %s\n" "$OUT_GENESIS"
printf "  -> %s\n" "$OUT_MIGRATIONS"
printf "  -> %s (kept for audit)\n" "$OUT_FORK_TEST_LOG"
