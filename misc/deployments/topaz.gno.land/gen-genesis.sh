#!/usr/bin/env bash
# gen-genesis.sh — topaz genesis builder (single-file pipeline).
#
# topaz is a FRESH chain — no hardfork, no historical replay. Earlier
# iterations built it as a gnoland1 hardfork that also replayed test13's
# tx history; that design was abandoned (multi-GB genesis, hours of
# InitChain replay per node, endless tx patching — see README.md). This
# script builds the whole genesis from the repo's examples/ tree plus a
# handful of bootstrap txs, in minutes.
#
# What the genesis contains:
#
#   1. The FILTERED_PACKAGES example set (resolved with transitive deps),
#      addpkg'd by the deterministic GenesisDeployer key.
#   2. A bootstrap MsgRun (transactions/base/bootstrap/) that seeds the
#      sole GovDAO T1 member and locks dao.UpdateImpl's AllowedDAOs to
#      r/gov/dao/v3/impl. No transfer lock — topaz is unrestricted.
#   3. A names.Enable MsgCall (transactions/migration/names-enable/) so
#      namespace enforcement is on from genesis. Enable is gated on the
#      admin address hardcoded in r/sys/names/verifier.gno; the tx's
#      caller field is jq-patched to that address post-sign, which the
#      chain trusts under --skip-genesis-sig-verification.
#   4. Two valopers.Register MsgCalls (emitted by `gnogenesis fork
#      valoper-seed` from INITIAL_VALSET + INITIAL_VALSET_OPERATORS) so
#      the founding validators have operator-keyed valoper profiles and
#      r/sys/validators/v3 can manage the set post-genesis.
#   5. The INITIAL_VALSET as GenesisDoc.Validators (InitChainer seeds
#      valset:current from it, so v3/EndBlocker valset changes work).
#   6. Balances: the 10 faucets at FAUCET_BALANCE each, plus exact-burn
#      funding for every genesis-tx fee payer (measured on a temp node;
#      those accounts land at zero once the genesis txs execute).
#
# Output:
#   work/genesis_txs.jsonl   full genesis tx stream (audit artifact)
#   work/valoper-seed.jsonl  valoper Register txs (audit artifact)
#   genesis.json             final artifact, sha256-locked against the
#                            CHECKSUMS_DATA heredoc in this script
#
# Usage:
#   ./gen-genesis.sh                # full build
#   ./gen-genesis.sh --debug        # echo every external command
#   ./gen-genesis.sh --no-install   # reuse previously built binaries
#
# Cross-platform: bash 3.2 minimum (macOS default), no GNU-only features.

set -eo pipefail

# =============================================================================
# Launch parameters — review before each genesis generation.
# =============================================================================

CHAIN_ID=topaz-1
GENESIS_TIME=1783868400 # Sunday, July 12th 2026 17:00 CEST (15:00 UTC)

# Packages to include in genesis (resolved with transitive dependencies).
# Use "..." suffix to match all sub-packages.
#
# First seven lines mirror gnoland1's gen-genesis.sh FILTERED_PACKAGES. The
# last block is additions carried over from test13:
#   - p/onbloc/{uint256,int256,json}: used by realms we want available
#     (uint256 is a transitive dep of int256).
#   - r/sys/validators/v3: the valset realm. The node's EndBlocker reads
#     valset state from this realm's params; without it on chain,
#     post-genesis valset changes can't happen.
#   - r/demo/defi/grc20reg: GRC20 token registry.
FILTERED_PACKAGES=(
  ./gno.land/r/sys/...
  ./gno.land/r/gov/...
  ./gno.land/r/gnoland/blog/...
  ./gno.land/r/gnoland/wugnot/...
  ./gno.land/r/gnoland/coins/...
  ./gno.land/r/gnoland/boards2/...
  ./gno.land/r/gnops/valopers/...
  ./gno.land/p/onbloc/uint256
  ./gno.land/p/onbloc/int256
  ./gno.land/p/onbloc/json
  ./gno.land/r/sys/validators/v3
  ./gno.land/r/demo/defi/grc20reg
)

# Initial topaz validator set. Format: "name power address pub_key".
# Power 60 each (cosmetic — consensus is about ratios, not absolutes).
INITIAL_VALSET=(
  "gno-core-val-01 60 g1pxtunv92xre6vsljecadrvqnwjwl9eyp63v32k gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqay44wamel8f4zvpjkfzfzpt36xwt8jg8r7zjkfj8x8rx7j6ekwq9ah2r9"
  "gno-core-val-02 60 g1zm0jtkxd4kz8jgkn03a0ggc3uax3epy2xp7urh gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp9m3ga2acvcdk04nrzgezuqc604w0jet7ql2jj0n7rxg2hgcs964ccmwxn"
)

# Operator address for each INITIAL_VALSET entry (same index). MUST be
# distinct from the signing address — `gnogenesis fork valoper-seed`
# rejects operator==signing_addr to keep signing-key compromise from
# collapsing into operator-slot compromise (see valoper_seed.go).
#
# The operator key is the management plane for the validator: whoever
# holds it can rotate the signing key, edit the valoper profile, and
# signal opt-out via r/gnops/valopers + r/sys/validators/v3.
INITIAL_VALSET_OPERATORS=(
  "g18x425qmujg99cfz3q97y4uep5pxjq3z8lmpt25" # gno-core-val-01 operator
  "g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m" # gno-core-val-02 operator
)

# Faucet balances. Each gets $FAUCET_BALANCE ugnot at genesis. Addresses
# are pasted from `gnokey list` output of an off-tree keybase (mnemonics
# are NOT in this repo). Transfers are unrestricted on topaz, so no
# unrestrict step is needed anywhere.
FAUCET_BALANCE=1000000000000000000 # 1e18 ugnot per faucet (1 trillion GNOT) — ~9.2x headroom under int64 max
FAUCET_ADDRESSES=(
  g15tjaykkykxa7e8nmtagm2swkphchj4j6rnukes # faucet-01
  g18qhq2fl54lszhmxeyqlvxnwjzc3xpu4nnakclp # faucet-02
  g18kre0dtu9sz25ux67pgcjfdqhas525rls34xz9 # faucet-03
  g157heusxh73m0wh6myjfd2f69uuwuu77kcc9vhs # faucet-04
  g18x40r2smn4telaps0cg9cw2znsjhnay9d353qh # faucet-05
  g16pdtpgrcwtq0hvh5lvdlrffx72pf8exqd4pnzn # faucet-06
  g18tv2p7jyk8dfqwl07v3vnyarvdz7ggprvlq8kt # faucet-07
  g1c24pc2rt3clps6tyd97rtsfevluxk5lp9k8n6e # faucet-08
  g1l58jdp5yannfd027j6yq37hprpkv32lnadshcm # faucet-09
  g1k28nhw04v54602jkdfrnu25gq07nyc2rehz9vl # faucet-10
)

# =============================================================================
# Internal — everything below is glue, you shouldn't need to change it.
# =============================================================================

# Deployer key mnemonic (deterministic — used only to sign genesis-mode txs).
# Same as gnoland1/test13 so the deployer address is reproducible.
DEPLOYER_MNEMONIC="anchor hurt name seed oak spread anchor filter lesson shaft wasp home improve text behind toe segment lamp turn marriage female royal twice wealth"
DEPLOYER_KEY=GenesisDeployer
# Address derived from DEPLOYER_MNEMONIC. Used as the fee payer for the
# valoper-seed Register txs; the balance-measurement step funds it exactly.
DEPLOYER_ADDR=g1edq4dugw0sgat4zxcw9xardvuydqf6cgleuc8p

# r/sys/names admin: hardcoded in examples/gno.land/r/sys/names/verifier.gno
# (the gnoland1 GovDAO T1 multisig). names.Enable's admin check reads
# runtime.PreviousRealm().Address(); under --skip-genesis-sig-verification,
# the chain trusts the MsgCall.Caller field as the EOA, so jq-patching
# caller to this address makes Enable's gate pass. The private key is not
# needed (and not held).
NAMES_ADMIN=g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh

# ---- Locked sha256 hashes.
#
# Format (matches `shasum -a 256` / `sha256sum` output exactly):
#   <sha256>  <path-relative-to-topaz.gno.land>
#
# Two spaces between hash and path. Blank lines and `#`-prefixed lines are
# ignored. The script calls `verify_checksum <path>` after producing an
# artifact:
#
#   - listed + hash matches  → silent pass
#   - listed + hash differs  → fail, expected vs got printed
#   - not listed             → note printed with the line to append
#
# Workflow: do a fresh end-to-end run, copy the "not listed" lines printed
# below this heredoc, commit, then any future run that produces a
# different output will fail loudly.
CHECKSUMS_DATA=$(
  cat <<'EOF'
# Build artifacts
dac29a0caae126dfbad0a2d5ad08e0aa5fb4f4c0da87b55bc8fb889c7ff20eb3  work/packages.gen.txt
c0946127b1ed0f310166c88f808c499df65f78f0c61f7032a6979cd5075397c4  work/valoper-seed.jsonl
a5aa589086ae36f8a74ac9f79fa7969d5b5c405637607e79d6d2565b28d2ba96  work/genesis_txs.jsonl

# Final artifact (moved to topaz.gno.land/ root on success)
2dd049f973b82858727440df9aff5722cb0b322fd00890f40f2b0688276898ff  genesis.json
EOF
)

# =============================================================================
# Helper functions.
# =============================================================================

# ---- Fatal error reporter

die() {
  printf 'ERROR: %s\n' "$*" >&2
  exit 1
}

# ---- Tool dispatchers

# sha256_of <path>
# Prints lowercase hex sha256 of the file's content. Tries shasum (macOS +
# most Linux), falls back to sha256sum (some Linux distros without shasum).
sha256_of() {
  local path="$1"
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$path" | awk '{print $1}'
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$path" | awk '{print $1}'
  else
    die "neither shasum nor sha256sum is installed (need one of them)"
  fi
}

# ---- Tool preflight
# require_tools <tool>...
# Probes every named tool; if any are missing, prints the full list with
# install hints (apt + brew) and exits. "shasum|sha256sum" is an
# at-least-one group; every other name is checked independently.
require_tools() {
  local missing=""
  local tool
  for tool in "$@"; do
    case "$tool" in
    "shasum|sha256sum")
      if ! command -v shasum >/dev/null 2>&1 && ! command -v sha256sum >/dev/null 2>&1; then
        missing="$missing shasum|sha256sum"
      fi
      ;;
    *)
      if ! command -v "$tool" >/dev/null 2>&1; then
        missing="$missing $tool"
      fi
      ;;
    esac
  done

  if [ -z "$missing" ]; then
    return 0
  fi

  printf 'ERROR: missing required tools:\n' >&2
  local m
  for m in $missing; do
    printf '  - %s\n' "$m" >&2
    case "$m" in
    "shasum|sha256sum")
      printf '      install:  brew install coreutils   |   apt-get install -y coreutils\n' >&2
      ;;
    jq)
      printf '      install:  brew install jq   |   apt-get install -y jq\n' >&2
      ;;
    go)
      printf '      install:  brew install go   |   see https://go.dev/doc/install\n' >&2
      ;;
    python3)
      printf '      install:  brew install python3   |   apt-get install -y python3\n' >&2
      ;;
    awk | sed | grep | sort | tr | mv | cp | ls | find | wc | head | tail | cut)
      printf '      install:  comes with any POSIX userland (coreutils + findutils)\n' >&2
      ;;
    *)
      printf '      install:  consult your package manager\n' >&2
      ;;
    esac
  done
  exit 1
}

# ---- Checksum verification
# verify_checksum <path>
#
# Computes sha256 of the file at <path>, looks up <path> (relative to
# TOPAZ_DIR) in the inline CHECKSUMS_DATA heredoc, and one of:
#   - hash matches               → silent OK
#   - hash differs               → FAIL with expected vs got
#   - path not listed            → print computed sha256 + the line to append
verify_checksum() {
  local path="$1"
  if [ -z "${TOPAZ_DIR:-}" ]; then
    die "verify_checksum: TOPAZ_DIR not set"
  fi
  if [ ! -f "$path" ]; then
    die "verify_checksum: $path does not exist"
  fi

  local rel="${path#"$TOPAZ_DIR"/}"
  local got
  got=$(sha256_of "$path")

  local expected
  expected=$(printf '%s\n' "$CHECKSUMS_DATA" | awk -v rel="$rel" '
    /^[[:space:]]*$/ { next }
    /^[[:space:]]*#/ { next }
    {
      if ($2 == rel) { print $1; exit }
    }
  ')

  if [ -z "$expected" ]; then
    printf '  [checksum] %s\n' "$rel" >&2
    printf '             not listed in CHECKSUMS_DATA. Append to lock:\n' >&2
    printf '             %s  %s\n' "$got" "$rel" >&2
    return 0
  fi

  if [ "$expected" = "$got" ]; then
    return 0
  fi

  printf 'ERROR: checksum mismatch for %s\n' "$rel" >&2
  printf '       expected: %s\n' "$expected" >&2
  printf '       got:      %s\n' "$got" >&2
  exit 1
}

# ---- Output helpers
# print_step_header <step> <total> <title>
#   prints e.g. `=== Step 3 of 9: Build binaries from source ===`
print_step_header() {
  local step="$1"
  local total="$2"
  local title="$3"
  printf '\n=== Step %s of %s: %s ===\n' "$step" "$total" "$title"
}

# print_substep <code> <text>
#   prints e.g. `  [3.1] Building gno...`
print_substep() {
  local code="$1"
  shift
  printf '  [%s] %s\n' "$code" "$*"
}

# ---- Formatting helpers

# format_duration <seconds>
# Prints "<H> hours <M> minutes <S> seconds" with zero parts omitted.
format_duration() {
  local s="$1"
  if [ "$s" -lt 0 ]; then s=0; fi
  local h=$((s / 3600))
  local m=$(((s % 3600) / 60))
  local sec=$((s % 60))
  local out=""
  if [ "$h" -gt 0 ]; then out="$h hours"; fi
  if [ "$m" -gt 0 ]; then
    if [ -n "$out" ]; then out="$out "; fi
    out="${out}$m minutes"
  fi
  if [ "$sec" -gt 0 ] || [ -z "$out" ]; then
    if [ -n "$out" ]; then out="$out "; fi
    out="${out}$sec seconds"
  fi
  printf '%s' "$out"
}

# format_size <bytes>
# Prints "245 MB", "4 KB", "789 B". Decimal units (1000-based).
format_size() {
  local b="$1"
  if [ "$b" -ge 1000000000 ]; then
    awk -v b="$b" 'BEGIN { printf "%.1f GB", b/1000000000 }'
  elif [ "$b" -ge 1000000 ]; then
    awk -v b="$b" 'BEGIN { printf "%.0f MB", b/1000000 }'
  elif [ "$b" -ge 1000 ]; then
    awk -v b="$b" 'BEGIN { printf "%.0f KB", b/1000 }'
  else
    printf '%s B' "$b"
  fi
}

# file_size <path>  →  bytes (uses wc -c, which is portable; stat flags differ)
file_size() {
  wc -c <"$1" | tr -d ' '
}

# =============================================================================
# Flag parsing.
# =============================================================================

DEBUG=false
NO_INSTALL=false

print_usage() {
  cat <<'EOF'
gen-genesis.sh — topaz genesis builder (single-file pipeline).

Usage:
  ./gen-genesis.sh [flags]

Flags:
  --no-install    Reuse previously built binaries in work/bin/.
  --debug         Echo every external command before running it.
  -h, --help      Print this help and exit.

Output:
  genesis.json    Final artifact, sha256-locked against the
                  CHECKSUMS_DATA heredoc in this script.

See misc/deployments/topaz.gno.land/README.md for what the genesis
contains and why topaz is a fresh chain rather than a hardfork.
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
  -h | --help)
    print_usage
    exit 0
    ;;
  --debug)
    DEBUG=true
    shift
    ;;
  --no-install)
    NO_INSTALL=true
    shift
    ;;
  *)
    echo "ERROR: Unknown argument: $1" >&2
    echo "Run with --help for usage." >&2
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

# =============================================================================
# Shared paths + cleanup trap.
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOPAZ_DIR="$SCRIPT_DIR"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
EXAMPLES_DIR="$REPO_ROOT/examples"

WORK_DIR="$SCRIPT_DIR/work"
WORK_DIR_BIN="$WORK_DIR/bin"
WORK_DIR_GNOKEY_HOME="$WORK_DIR/gnokey-home"

GNO_CMD="$REPO_ROOT/gnovm/cmd/gno"
GNOKEY_CMD="$REPO_ROOT/gno.land/cmd/gnokey"
GNOLAND_CMD="$REPO_ROOT/gno.land/cmd/gnoland"
GNOGENESIS_CMD="$REPO_ROOT/contribs/gnogenesis"
GNO_BIN="$WORK_DIR_BIN/gno"
GNOKEY_BIN="$WORK_DIR_BIN/gnokey"
GNOLAND_BIN="$WORK_DIR_BIN/gnoland"
GNOGENESIS_BIN="$WORK_DIR_BIN/gnogenesis"

FINAL_GENESIS="$SCRIPT_DIR/genesis.json"

# Clean up temp node on exit (the balance-measurement step starts one; the
# trap is a no-op when NODE_PID is unset, so it's safe at script scope).
NODE_PID=""
cleanup() { [ -n "$NODE_PID" ] && kill "$NODE_PID" 2>/dev/null || true; }
trap cleanup EXIT

# =============================================================================
# Transaction loader — converts transactions/<...>/<txdir>/{meta.json +
# optional body file} into one AnnotatedTx jsonl line appended to <outfile>.
#
# Dispatches on meta.json's "kind" field:
#   MsgRun   — signs via gnokey maketx run + sign.
#   MsgCall  — signs via gnokey maketx call + sign; optionally jq-patches
#              msg[0].caller to caller_override post-sign (for genesis-mode
#              calls that need an admin caller without holding the admin
#              key — the chain trusts the caller field under
#              --skip-genesis-sig-verification).
#
# Emitted lines are in AnnotatedTx shape {tx, metadata, reason} — strip the
# reason field (jq -c 'del(.reason)') before passing to consumers that
# don't speak AnnotatedTx (e.g. `gnogenesis txs add sheets`).
# =============================================================================

txn_dir_to_jsonl() {
  local dir="$1" outfile="$2"
  local meta="$dir/meta.json"
  [ -f "$meta" ] || die "txn_dir_to_jsonl: $meta not found"

  local kind
  kind=$(jq -r '.kind' "$meta")
  case "$kind" in
  MsgRun) _txn_msg_run "$dir" "$outfile" ;;
  MsgCall) _txn_msg_call "$dir" "$outfile" ;;
  *) die "txn_dir_to_jsonl: unknown kind '$kind' in $meta" ;;
  esac
}

_txn_msg_run() {
  local dir="$1" outfile="$2"
  local meta="$dir/meta.json"

  local reason caller_key body_file gas_wanted gas_fee acct_num seq
  reason=$(jq -r '.reason' "$meta")
  caller_key=$(jq -r '.caller_key' "$meta")
  body_file=$(jq -r '.body_file' "$meta")
  gas_wanted=$(jq -r '.gas_wanted' "$meta")
  gas_fee=$(jq -r '.gas_fee' "$meta")
  acct_num=$(jq -r '.account_number' "$meta")
  seq=$(jq -r '.sequence' "$meta")

  local body_path="$dir/$body_file"
  local tx_json="$dir/.txn.json"

  run "$GNOKEY_BIN" maketx run \
    --gas-wanted "$gas_wanted" \
    --gas-fee "$gas_fee" \
    --chainid "$CHAIN_ID" \
    --home "$WORK_DIR_GNOKEY_HOME" \
    --broadcast=false \
    --insecure-password-stdin \
    "$caller_key" \
    "$body_path" >"$tx_json" <<<""

  echo "" | run "$GNOKEY_BIN" sign \
    --tx-path "$tx_json" \
    --chainid "$CHAIN_ID" \
    --account-number "$acct_num" \
    --account-sequence "$seq" \
    --home "$WORK_DIR_GNOKEY_HOME" \
    --insecure-password-stdin \
    "$caller_key" >/dev/null

  jq -c --arg r "$reason" '{tx: ., metadata: {block_height: "0"}, reason: $r}' \
    "$tx_json" >>"$outfile"
  rm -f "$tx_json"
}

_txn_msg_call() {
  local dir="$1" outfile="$2"
  local meta="$dir/meta.json"

  local reason caller_key caller_override pkgpath func gas_wanted gas_fee acct_num seq
  reason=$(jq -r '.reason' "$meta")
  caller_key=$(jq -r '.caller_key' "$meta")
  caller_override=$(jq -r '.caller_override // empty' "$meta")
  pkgpath=$(jq -r '.pkgpath' "$meta")
  func=$(jq -r '.func' "$meta")
  gas_wanted=$(jq -r '.gas_wanted' "$meta")
  gas_fee=$(jq -r '.gas_fee' "$meta")
  acct_num=$(jq -r '.account_number' "$meta")
  seq=$(jq -r '.sequence' "$meta")

  # Expand args: meta.json's .args is a JSON array of strings; pass each as --args.
  local args_array=() arg
  while IFS= read -r arg; do
    [ -z "$arg" ] && continue
    args_array+=(--args "$arg")
  done < <(jq -r '.args[]?' "$meta")

  local tx_json="$dir/.txn.json"

  echo "" | run "$GNOKEY_BIN" maketx call \
    --pkgpath "$pkgpath" \
    --func "$func" \
    "${args_array[@]}" \
    --gas-wanted "$gas_wanted" \
    --gas-fee "$gas_fee" \
    --chainid "$CHAIN_ID" \
    --home "$WORK_DIR_GNOKEY_HOME" \
    --broadcast=false \
    --insecure-password-stdin \
    "$caller_key" >"$tx_json"

  echo "" | run "$GNOKEY_BIN" sign \
    --tx-path "$tx_json" \
    --chainid "$CHAIN_ID" \
    --account-number "$acct_num" \
    --account-sequence "$seq" \
    --home "$WORK_DIR_GNOKEY_HOME" \
    --insecure-password-stdin \
    "$caller_key" >/dev/null

  if [ -n "$caller_override" ]; then
    jq -c --arg c "$caller_override" --arg r "$reason" \
      '.msg[0].caller = $c | {tx: ., metadata: {block_height: "0"}, reason: $r}' \
      "$tx_json" >>"$outfile"
  else
    jq -c --arg r "$reason" \
      '{tx: ., metadata: {block_height: "0"}, reason: $r}' \
      "$tx_json" >>"$outfile"
  fi
  rm -f "$tx_json"
}

# =============================================================================
# Pipeline.
# =============================================================================

PIPELINE_START_TS=$(date +%s)
TOTAL_STEPS=9

printf '\n### topaz genesis build ###\n'

# ---- Step 1: Resolve script paths and tooling

print_step_header 1 "$TOTAL_STEPS" "Resolve script paths and tooling"

GENESIS_FILE="$WORK_DIR/genesis.json"
PACKAGES_GEN_FILE="$WORK_DIR/packages.gen.txt"
GENESIS_TXS_JSONL="$WORK_DIR/genesis_txs.jsonl"
DEPLOYER_BALANCES="$WORK_DIR/deployers_balances.txt"
VALOPER_CSV="$WORK_DIR/valoper_profiles.csv"
VALOPER_SEED="$WORK_DIR/valoper-seed.jsonl"

print_substep "1.1" "TOPAZ_DIR=$TOPAZ_DIR"
print_substep "1.2" "REPO_ROOT=$REPO_ROOT"
print_substep "1.3" "WORK_DIR=$WORK_DIR"

# ---- Step 2: Verify required tools

print_step_header 2 "$TOTAL_STEPS" "Verify required tools"

require_tools \
  "shasum|sha256sum" \
  go jq python3 \
  awk sed grep sort tr mv cp ls find wc head tail cut

print_substep "2.1" "All required tools present"

# Prepare work dir; preserve bin/ when --no-install.
if [ "$NO_INSTALL" = true ]; then
  mkdir -p "$WORK_DIR"
  find "$WORK_DIR" -mindepth 1 -maxdepth 1 ! -name bin -exec rm -rf {} + 2>/dev/null || true
else
  rm -rf "$WORK_DIR"
fi
mkdir -p "$WORK_DIR_BIN"

# ---- Step 3: Build binaries from source

print_step_header 3 "$TOTAL_STEPS" "Build binaries from source"

if [ "$NO_INSTALL" = true ]; then
  print_substep "3.1" "--no-install — reusing $WORK_DIR_BIN"
  for bin in "$GNO_BIN" "$GNOKEY_BIN" "$GNOLAND_BIN" "$GNOGENESIS_BIN"; do
    if [ ! -x "$bin" ]; then
      die "--no-install but $bin not found. Run without --no-install first."
    fi
  done
else
  print_substep "3.1" "Building gno..."
  run go build -C "$GNO_CMD" -o "$GNO_BIN" .
  print_substep "3.2" "Building gnokey..."
  run go build -C "$GNOKEY_CMD" -o "$GNOKEY_BIN" .
  print_substep "3.3" "Building gnoland..."
  run go build -C "$GNOLAND_CMD" -o "$GNOLAND_BIN" .
  print_substep "3.4" "Building gnogenesis..."
  run go build -C "$GNOGENESIS_CMD" -o "$GNOGENESIS_BIN" .
fi

# ---- Step 4: Generate filtered examples genesis txs

print_step_header 4 "$TOTAL_STEPS" "Generate filtered examples genesis txs"

print_substep "4.1" "Resolving dependencies..."
pkg_dirs=$(cd "$EXAMPLES_DIR" && "$GNO_BIN" tool deplist -test-dep "${FILTERED_PACKAGES[@]}")
pkg_count=$(echo "$pkg_dirs" | wc -l | tr -d ' ')
print_substep "4.2" "Resolved $pkg_count packages in topological order"

# Save resolved package list (used for audit + tracked by CHECKSUMS).
{
  echo "# Generated by gen-genesis.sh — do not edit."
  # shellcheck disable=SC2001 # path contains slashes; `|` as sed delimiter is cleaner than ${//} escaping
  echo "$pkg_dirs" | sed "s|$EXAMPLES_DIR/||g"
} >"$PACKAGES_GEN_FILE"
verify_checksum "$PACKAGES_GEN_FILE"

print_substep "4.3" "Copying packages to staging..."
WORK_DIR_EXAMPLES="$WORK_DIR/examples"
mkdir -p "$WORK_DIR_EXAMPLES"
while IFS= read -r dir; do
  [ -z "$dir" ] && continue
  rel="${dir#"$EXAMPLES_DIR"/}"
  dest="$WORK_DIR_EXAMPLES/$rel"
  mkdir -p "$dest"
  find "$dir" -maxdepth 1 -type f -exec cp {} "$dest/" \;
  if [ -d "$dir/filetests" ]; then
    cp -r "$dir/filetests" "$dest/filetests"
  fi
done <<<"$pkg_dirs"

print_substep "4.4" "Creating deployer key..."
printf '%s\n\n' "$DEPLOYER_MNEMONIC" | run "$GNOKEY_BIN" add --recover "$DEPLOYER_KEY" --home "$WORK_DIR_GNOKEY_HOME" --insecure-password-stdin 2>&1 | sed 's/^/    /'

print_substep "4.5" "Generating empty genesis..."
run "$GNOGENESIS_BIN" generate -chain-id "$CHAIN_ID" -genesis-time "$GENESIS_TIME" --output-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'

print_substep "4.6" "Adding $pkg_count packages to genesis..."
echo "" | run "$GNOGENESIS_BIN" txs add packages "$WORK_DIR_EXAMPLES" -gno-home "$WORK_DIR_GNOKEY_HOME" -key-name "$DEPLOYER_KEY" --genesis-path "$GENESIS_FILE" --insecure-password-stdin 2>&1 | sed 's/^/    /'

print_substep "4.7" "Exporting txs..."
run "$GNOGENESIS_BIN" txs export "$GENESIS_TXS_JSONL" --genesis-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'

# ---- Step 5: Add the bootstrap MsgRun (transactions/base/bootstrap/)
# Seeds the sole GovDAO T1 member (aeddi) and locks AllowedDAOs. No
# transfer lock, no unrestricted-accounts proposals — topaz transfers
# are unrestricted.

print_step_header 5 "$TOTAL_STEPS" "Add bootstrap MsgRun (GovDAO seed)"

BOOTSTRAP_DIR="$SCRIPT_DIR/transactions/base/bootstrap"
BOOTSTRAP_JSONL="$WORK_DIR/bootstrap_tx.jsonl"

print_substep "5.1" "Building AnnotatedTx from $BOOTSTRAP_DIR/..."
: >"$BOOTSTRAP_JSONL"
txn_dir_to_jsonl "$BOOTSTRAP_DIR" "$BOOTSTRAP_JSONL"

# `txs add sheets` consumes plain TxWithMetadata (no reason field); strip it.
BOOTSTRAP_TX_FILE="$WORK_DIR/bootstrap_tx_stripped.jsonl"
jq -c 'del(.reason)' "$BOOTSTRAP_JSONL" >"$BOOTSTRAP_TX_FILE"

print_substep "5.2" "Adding bootstrap tx to genesis..."
run "$GNOGENESIS_BIN" txs add sheets "$BOOTSTRAP_TX_FILE" --genesis-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'
cat "$BOOTSTRAP_TX_FILE" >>"$GENESIS_TXS_JSONL"

# ---- Step 6: Add the names.Enable MsgCall (transactions/migration/names-enable/)
# Namespace enforcement on from genesis. Enable is gated on the admin
# hardcoded in r/sys/names/verifier.gno; caller_override makes the tx
# appear as that admin (trusted under --skip-genesis-sig-verification).
# Ordered AFTER every addpkg, so enforcement never gates the genesis
# deploys themselves.

print_step_header 6 "$TOTAL_STEPS" "Add names.Enable MsgCall (namespace enforcement)"

NAMES_ENABLE_DIR="$SCRIPT_DIR/transactions/migration/names-enable"
NAMES_ENABLE_JSONL="$WORK_DIR/names_enable_tx.jsonl"

print_substep "6.1" "Building AnnotatedTx from $NAMES_ENABLE_DIR/..."
: >"$NAMES_ENABLE_JSONL"
txn_dir_to_jsonl "$NAMES_ENABLE_DIR" "$NAMES_ENABLE_JSONL"

NAMES_ENABLE_TX_FILE="$WORK_DIR/names_enable_tx_stripped.jsonl"
jq -c 'del(.reason)' "$NAMES_ENABLE_JSONL" >"$NAMES_ENABLE_TX_FILE"

print_substep "6.2" "Adding names.Enable tx to genesis (caller=$NAMES_ADMIN)..."
run "$GNOGENESIS_BIN" txs add sheets "$NAMES_ENABLE_TX_FILE" --genesis-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'
cat "$NAMES_ENABLE_TX_FILE" >>"$GENESIS_TXS_JSONL"

# ---- Step 7: Add the valoper-seed Register MsgCalls
# Builds a CSV from (INITIAL_VALSET, INITIAL_VALSET_OPERATORS) and runs
# `gnogenesis fork valoper-seed` to produce a deterministic .jsonl of
# genesis-mode valopers.Register MsgCalls — one valoper profile per
# founding validator, keyed on its operator address. Without these the
# chain still boots (the valoper coverage assertion only fires in
# hardfork mode), but the founding validators would have no operator-
# keyed management plane in r/sys/validators/v3.

print_step_header 7 "$TOTAL_STEPS" "Add valoper-seed Register MsgCalls"

if [ "${#INITIAL_VALSET_OPERATORS[@]}" -ne "${#INITIAL_VALSET[@]}" ]; then
  die "INITIAL_VALSET_OPERATORS length (${#INITIAL_VALSET_OPERATORS[@]}) must match INITIAL_VALSET length (${#INITIAL_VALSET[@]})"
fi

print_substep "7.1" "Building CSV from INITIAL_VALSET + INITIAL_VALSET_OPERATORS..."
{
  echo "operator_addr,signing_pubkey,moniker,description,server_type"
  for i in "${!INITIAL_VALSET[@]}"; do
    read -r name _power _address pub_key <<<"${INITIAL_VALSET[$i]}"
    op_addr="${INITIAL_VALSET_OPERATORS[$i]}"
    # description and server_type are templates — edit if a specific
    # founder needs different metadata. Description must be non-empty
    # and <=2048 chars; server_type ∈ {cloud, on-prem, data-center}.
    printf '%s,%s,%s,topaz founding validator (%s),cloud\n' \
      "$op_addr" "$pub_key" "$name" "$name"
  done
} >"$VALOPER_CSV"

print_substep "7.2" "Running gnogenesis fork valoper-seed..."
# --caller is the fee payer for each Register MsgCall (1 ugnot fee — see
# the Coin amino zero-collapse rationale in valoper_seed.go). The deployer
# pays; the balance-measurement step (step 8) funds it exactly. The
# operator from the CSV row is passed in MsgCall.Args[3], so each operator
# gets registered correctly; the squat guard (caller==operator) is
# bypassed at genesis-mode (ChainHeight()==0).
run "$GNOGENESIS_BIN" fork valoper-seed \
  --csv "$VALOPER_CSV" \
  --output "$VALOPER_SEED" \
  --caller "$DEPLOYER_ADDR" 2>&1 | sed 's/^/    /'
verify_checksum "$VALOPER_SEED"

VALOPER_TX_FILE="$WORK_DIR/valoper_seed_stripped.jsonl"
jq -c 'del(.reason)' "$VALOPER_SEED" >"$VALOPER_TX_FILE"

print_substep "7.3" "Adding ${#INITIAL_VALSET[@]} valoper Register txs to genesis..."
run "$GNOGENESIS_BIN" txs add sheets "$VALOPER_TX_FILE" --genesis-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'
cat "$VALOPER_TX_FILE" >>"$GENESIS_TXS_JSONL"
verify_checksum "$GENESIS_TXS_JSONL"

tx_count=$(wc -l <"$GENESIS_TXS_JSONL" | tr -d ' ')
print_substep "7.4" "Total genesis txs: $tx_count"

# ---- Step 8: Calculate genesis fee-payer balances
# Same approach as gnoland1's gen-genesis.sh: spin up a temp node, pre-fund
# every creator/caller address with $INITIAL_BALANCE, let the genesis txs
# burn through fees, then query remaining balances. The amount actually
# spent is what we credit each fee payer in the real genesis so their
# balance lands at zero post-genesis — the final state then holds ONLY the
# 10 faucet balances. This covers the deployer (addpkgs, bootstrap,
# valoper Register fees) and the names admin (names.Enable fee).
#
# Run twice for safety:
#   run 1: measure actual consumption with over-provisioned balances
#   run 2: verify the measured balances land everyone at zero
# If run 2 disagrees, something is non-deterministic and we abort.

print_step_header 8 "$TOTAL_STEPS" "Calculate genesis fee-payer balances"

BALANCES_TMP_DIR="$WORK_DIR/balances-work"
BALANCES_TMP_FILE="$BALANCES_TMP_DIR/balances.txt"
BALANCES_TMP_GNOLAND_DATA="$BALANCES_TMP_DIR/gnoland-data"
BALANCES_TMP_GNOLAND_LOG="$BALANCES_TMP_DIR/node.log"
BALANCES_TMP_GENESIS="$BALANCES_TMP_DIR/genesis.json"
BALANCES_TMP_CREATOR_ADDRESSES="$BALANCES_TMP_DIR/gen-creators.txt"
INITIAL_BALANCE=1000000000000000
NODE_TIMEOUT=120

pick_free_port() {
  python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()'
}

NODE_RPC_PORT=$(pick_free_port)
NODE_P2P_PORT=$((NODE_RPC_PORT + 1))
NODE_RPC_ADDR="127.0.0.1:$NODE_RPC_PORT"

rm -rf "$BALANCES_TMP_DIR"
mkdir -p "$BALANCES_TMP_DIR"

print_substep "8.1" "Extracting creator/caller addresses..."
grep -oE '"(creator|caller)":"[^"]*"' "$GENESIS_TXS_JSONL" |
  sed 's/"creator":"//;s/"caller":"//;s/"//g' |
  sort -u >"$BALANCES_TMP_CREATOR_ADDRESSES"
addr_count=$(wc -l <"$BALANCES_TMP_CREATOR_ADDRESSES" | tr -d ' ')
print_substep "8.2" "Found $addr_count unique creator/caller addresses"

print_substep "8.3" "Generating over-provisioned balances..."
while IFS= read -r addr; do
  echo "${addr}=${INITIAL_BALANCE}ugnot" >>"$BALANCES_TMP_FILE"
done <"$BALANCES_TMP_CREATOR_ADDRESSES"

# Helper: spin up a temp node with the current genesis + balance sheet.
# Sets NODE_PID; aborts if the node doesn't come up in NODE_TIMEOUT seconds.
start_temp_node() {
  local run_label="$1"
  rm -rf "$BALANCES_TMP_GNOLAND_DATA" "$BALANCES_TMP_GENESIS"
  NODE_RPC_PORT=$(pick_free_port)
  NODE_P2P_PORT=$((NODE_RPC_PORT + 1))
  NODE_RPC_ADDR="127.0.0.1:$NODE_RPC_PORT"

  run "$GNOGENESIS_BIN" generate -chain-id "$CHAIN_ID" -genesis-time "$(date +%s)" -output-path "$BALANCES_TMP_GENESIS"
  run "$GNOGENESIS_BIN" txs add sheets "$GENESIS_TXS_JSONL" -genesis-path "$BALANCES_TMP_GENESIS"
  run "$GNOGENESIS_BIN" balances add -balance-sheet "$BALANCES_TMP_FILE" -genesis-path "$BALANCES_TMP_GENESIS"
  run "$GNOLAND_BIN" config init -config-path "$BALANCES_TMP_GNOLAND_DATA/config/config.toml"
  run "$GNOLAND_BIN" config set rpc.laddr "tcp://$NODE_RPC_ADDR" -config-path "$BALANCES_TMP_GNOLAND_DATA/config/config.toml"
  run "$GNOLAND_BIN" config set p2p.laddr "tcp://127.0.0.1:$NODE_P2P_PORT" -config-path "$BALANCES_TMP_GNOLAND_DATA/config/config.toml"
  run "$GNOLAND_BIN" secrets init -data-dir "$BALANCES_TMP_GNOLAND_DATA/secrets"
  run "$GNOGENESIS_BIN" validator add \
    --address "$("$GNOLAND_BIN" secrets get validator_key.address --raw -data-dir "$BALANCES_TMP_GNOLAND_DATA/secrets")" \
    --pub-key "$("$GNOLAND_BIN" secrets get validator_key.pub_key --raw -data-dir "$BALANCES_TMP_GNOLAND_DATA/secrets")" \
    --name balance_generator \
    --power 1 \
    -genesis-path "$BALANCES_TMP_GENESIS"

  printf "  Starting node (%s)...\n" "$run_label"
  "$GNOLAND_BIN" start --skip-genesis-sig-verification -data-dir "$BALANCES_TMP_GNOLAND_DATA" -genesis "$BALANCES_TMP_GENESIS" >"$BALANCES_TMP_GNOLAND_LOG" 2>&1 &
  NODE_PID=$!

  local elapsed=0
  while [ "$elapsed" -lt "$NODE_TIMEOUT" ]; do
    if ! kill -0 "$NODE_PID" 2>/dev/null; then
      echo "ERROR: Node stopped unexpectedly. Last log lines:" >&2
      tail -20 "$BALANCES_TMP_GNOLAND_LOG" >&2
      exit 1
    fi
    if "$GNOKEY_BIN" query -remote "$NODE_RPC_ADDR" "auth/accounts/$DEPLOYER_ADDR" >/dev/null 2>&1; then
      printf "  Node ready (%ss)\n" "$elapsed"
      return
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  kill "$NODE_PID" 2>/dev/null || true
  echo "ERROR: Node did not start within ${NODE_TIMEOUT}s. Last log lines:" >&2
  tail -20 "$BALANCES_TMP_GNOLAND_LOG" >&2
  exit 1
}

stop_temp_node() {
  kill "$NODE_PID" 2>/dev/null || true
  wait "$NODE_PID" 2>/dev/null || true
  NODE_PID=""
}

# query_balance ADDR → echoes ugnot amount (0 if empty)
query_balance() {
  local addr="$1"
  local retry=0
  while [ "$retry" -lt "$NODE_TIMEOUT" ]; do
    if ! kill -0 "$NODE_PID" 2>/dev/null; then
      echo "ERROR: Node stopped unexpectedly during balance query. Last log lines:" >&2
      tail -20 "$BALANCES_TMP_GNOLAND_LOG" >&2
      exit 1
    fi
    local out
    out=$("$GNOKEY_BIN" query -remote "$NODE_RPC_ADDR" "bank/balances/$addr" 2>&1 || true)
    if echo "$out" | grep -q '^data:'; then
      local r
      r=$(echo "$out" | sed -n 's/.*"\([0-9]*\)ugnot".*/\1/p' | head -1)
      echo "${r:-0}"
      return
    fi
    sleep 1
    retry=$((retry + 1))
  done
  echo "ERROR: Could not query balance for $addr after ${NODE_TIMEOUT}s." >&2
  exit 1
}

start_temp_node "run 1: measure gas costs"
print_substep "8.4" "Querying remaining balances..."
rm -f "$BALANCES_TMP_FILE"
while IFS= read -r addr; do
  remaining=$(query_balance "$addr")
  final=$((INITIAL_BALANCE - remaining))
  printf "    %s = %s ugnot\n" "$addr" "$final"
  echo "${addr}=${final}ugnot" >>"$BALANCES_TMP_FILE"
done <"$BALANCES_TMP_CREATOR_ADDRESSES"
stop_temp_node

start_temp_node "run 2: verify zero balances"
print_substep "8.5" "Verifying all balances are zero..."
all_zero=true
while IFS= read -r addr; do
  remaining=$(query_balance "$addr")
  if [ "$remaining" -ne 0 ]; then
    printf "    FAIL: %s has %sugnot remaining\n" "$addr" "$remaining"
    all_zero=false
  else
    printf "    ok: %s\n" "$addr"
  fi
done <"$BALANCES_TMP_CREATOR_ADDRESSES"
stop_temp_node

if [ "$all_zero" != true ]; then
  die "Some balances are not zero after replay. Check $BALANCES_TMP_FILE."
fi
print_substep "8.6" "All balances zero — fee-payer costs verified"
cp "$BALANCES_TMP_FILE" "$DEPLOYER_BALANCES"

# ---- Step 9: Add validators + balances, verify, move into place

print_step_header 9 "$TOTAL_STEPS" "Add validators + balances, verify genesis"

print_substep "9.1" "Adding the initial validator set..."
for validator in "${INITIAL_VALSET[@]}"; do
  read -r name power address pub_key <<<"$validator"
  printf "    %s (power=%s, %s)\n" "$name" "$power" "$address"
  run "$GNOGENESIS_BIN" validator add -name "$name" -power "$power" -address "$address" -pub-key "$pub_key" --genesis-path "$GENESIS_FILE"
done

# Fee payers (exact-burn, land at zero) + the 10 faucets, one sheet.
# No address can appear in both lists: fee payers are the deployer + the
# names admin, neither of which is a faucet.
FULL_BALANCES_FILE="$WORK_DIR/balances.txt"
cp "$DEPLOYER_BALANCES" "$FULL_BALANCES_FILE"
for addr in "${FAUCET_ADDRESSES[@]}"; do
  echo "${addr}=${FAUCET_BALANCE}ugnot" >>"$FULL_BALANCES_FILE"
done
balance_count=$(wc -l <"$FULL_BALANCES_FILE" | tr -d ' ')
print_substep "9.2" "Adding $balance_count balances (fee payers + ${#FAUCET_ADDRESSES[@]} faucets)..."
run "$GNOGENESIS_BIN" balances add -balance-sheet "$FULL_BALANCES_FILE" --genesis-path "$GENESIS_FILE" >/dev/null

print_substep "9.3" "Running gnogenesis verify..."
# -skip-signature-check: the names.Enable tx carries a post-sign caller
# patch and the valoper Register txs carry placeholder signatures, so
# per-tx signature verification cannot pass by design (nodes accept both
# under --skip-genesis-sig-verification). Every other verify check runs.
run "$GNOGENESIS_BIN" verify -genesis-path "$GENESIS_FILE" -skip-signature-check

print_substep "9.4" "Moving $GENESIS_FILE -> $FINAL_GENESIS"
mv "$GENESIS_FILE" "$FINAL_GENESIS"
verify_checksum "$FINAL_GENESIS"

# ---- Summary

PIPELINE_END_TS=$(date +%s)
PIPELINE_DURATION=$((PIPELINE_END_TS - PIPELINE_START_TS))
FINAL_SHA=$(sha256_of "$FINAL_GENESIS")
FINAL_BYTES=$(file_size "$FINAL_GENESIS")

printf '\n### topaz build complete: genesis.json (%s, sha256=%s) ###\n' \
  "$(format_size "$FINAL_BYTES")" "$FINAL_SHA"
printf '    total pipeline time: %s\n' "$(format_duration "$PIPELINE_DURATION")"
