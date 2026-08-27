#!/usr/bin/env bash
# gen-genesis.sh — pearl genesis builder (single-file pipeline).
#
# pearl is a FRESH chain — no hardfork, no historical replay. This
# script builds the whole genesis from the repo's examples/ tree plus a
# handful of bootstrap txs, in minutes.
#
# What the genesis contains:
#
#   1. The FILTERED_PACKAGES example set (resolved with transitive deps),
#      addpkg'd by the deterministic GenesisDeployer key.
#   2. A bootstrap MsgRun (transactions/base/bootstrap/) that seeds the
#      sole GovDAO T1 member and locks dao.UpdateImpl's AllowedDAOs to
#      r/gov/dao/v3/impl. No transfer lock — pearl is unrestricted.
#   3. A names.Enable MsgCall (transactions/migration/names-enable/) so
#      namespace enforcement is on from genesis. Enable is gated on the
#      admin address hardcoded in r/sys/names/verifier.gno; the tx's
#      caller field is jq-patched to that address post-sign, which the
#      chain trusts under --skip-genesis-sig-verification.
#   4. Per-validator valopers.Register MsgCalls (emitted by `gnogenesis fork
#      valoper-seed` from INITIAL_VALSET + INITIAL_VALSET_OPERATORS) so
#      the founding validators have operator-keyed valoper profiles and
#      r/sys/validators/v3 can manage the set post-genesis.
#   5. The INITIAL_VALSET as GenesisDoc.Validators (InitChainer seeds
#      valset:current from it, so v3/EndBlocker valset changes work).
#   6. Balances: the faucet accounts at FAUCET_BALANCE each, the
#      VESTED_ACCOUNTS entries (created as vesting accounts at genesis),
#      plus exact-burn funding for every genesis-tx fee payer (measured on
#      a temp node; those accounts land at zero once the genesis txs
#      execute).
#
# Output:
#   work/packages.gen.txt    resolved package list (audit artifact)
#   work/genesis_txs.jsonl   full genesis tx stream (audit artifact)
#   work/valoper-seed.jsonl  valoper Register txs (audit artifact)
#   genesis.json             final artifact, sha256-locked against the
#                            CHECKSUMS_DATA heredoc in this script
#
# Usage:
#   ./gen-genesis.sh                # full build
#   ./gen-genesis.sh --debug        # echo the main pipeline commands
#   ./gen-genesis.sh --no-install   # reuse previously built binaries
#
# Cross-platform: bash 3.2 minimum (macOS default), no GNU-only features.

set -eo pipefail

# =============================================================================
# Launch parameters — review before each genesis generation.
# =============================================================================

CHAIN_ID=pearl-1
GENESIS_TIME=1787752800 # Wednesday, August 26th 2026 16:00 CEST (14:00 UTC)

# Packages to include in genesis (resolved with transitive dependencies).
# Use "..." suffix to match all sub-packages.
#
# First seven lines mirror gnoland1's gen-genesis.sh FILTERED_PACKAGES. The
# last block is additions carried over from test13:
#   - p/onbloc/{uint256,int256,json}: used by realms we want available
#     (uint256 is a transitive dep of int256).
#   - r/sys/validators/v3: the valset realm — already matched by the
#     ./gno.land/r/sys/... pattern, kept explicit because it is load-
#     bearing: the node's EndBlocker reads valset state from this realm's
#     params; without it on chain, post-genesis valset changes can't
#     happen.
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

# Initial pearl validator set. Format: "name power address pub_key".
# Power 60 each — 60 divides evenly many ways (2, 3, 4, 5, 6, 10, ...),
# so later valset changes can hand out proportional fractions of the
# founders' voting power without fractional remainders.
#
# Three founding validators is an operational bootstrap, not the target
# topology: at 3 × 60, one validator dark loses exactly one third of the
# voting power — the halt boundary — accepted for launch. Partner joins
# post-genesis relax it, via r/gnops/valopers + r/sys/validators/v3.
#
# Validators 1/2 sign via horcrux clusters (2-of-3 threshold), validator
# 3 via tmkms softsign; the consensus keys never touch the validator
# nodes.
#
# Ceremony keys of 2026-08-24 (FINAL, from PEARL-PR-HANDOFF.md):
# validators 1/2 on horcrux clusters, validator 3 on tmkms.
INITIAL_VALSET=(
  "gno-core-validator-1 60 g1l92r5ystf0c97j923wn6dnvswjqnzwmyjtnpyg gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq6yf7gyvwpw8yywe9n6nwahp9g3s7wyydlqj0aakasx8r5twdqy0g245ue"
  "gno-core-validator-2 60 g1etztcjetpm5yuvsw2xdw0uvkqg0ur6tjzgtxtf gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqthqhdmvlzvv07xrfqmmtjyshlsnctt44wml26gnvh6pjh8wynml2jjjf3"
  "gno-core-validator-3 60 g1ca5dqn0h4nsk0sp95wzpswvacnmtzq2pqly765 gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp7qqkm9qz6g8u3nlc4pqlzau7pndrx99mar2vu6m0p92xxwwsutsr2qx9j"
)

# Operator address for each INITIAL_VALSET entry (same index). MUST be
# distinct from the signing address — `gnogenesis fork valoper-seed`
# rejects operator==signing_addr to keep signing-key compromise from
# collapsing into operator-slot compromise (see valoper_seed.go).
#
# The operator key is the management plane for the validator: whoever
# holds it can rotate the signing key, edit the valoper profile, and
# signal opt-out via r/gnops/valopers + r/sys/validators/v3.
#
# Valoper profiles are keyed on the operator address and `fork
# valoper-seed` rejects duplicate operators, so the three slots must be
# distinct addresses.
INITIAL_VALSET_OPERATORS=(
  "g18x425qmujg99cfz3q97y4uep5pxjq3z8lmpt25" # gno-core-validator-1 operator
  "g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m" # gno-core-validator-2 operator
  "g18e69gfvt9c5mw40ykfzvhtjlmur8guhp4zjtav" # gno-core-validator-3 operator
)

# Faucet balances. Each gets $FAUCET_BALANCE ugnot at genesis. Addresses
# are pasted from `gnokey list` output of an off-tree keybase (mnemonics
# are NOT in this repo). Transfers are unrestricted on pearl, so no
# unrestrict step is needed anywhere.
#
# 1e18 leaves ~9.2x headroom under int64 max per account. Keep the total
# genesis supply — (count × FAUCET_BALANCE) plus every VESTED_ACCOUNTS
# total — under int64 max (~9.22e18): Coin.Add panics on int64 overflow,
# so consolidating all genesis funds into one account must not be able to
# exceed it.
FAUCET_BALANCE=1000000000000000000 # 1e18 ugnot per faucet (1 trillion GNOT)
# Confirmed by PEARL-PR-HANDOFF.md: reuse sapphire's accounts and
# amounts. The faucet account doubles as the infra's snapshot-validation
# probe (gnoland_snapshot_probe_address), so funding it is load-bearing.
FAUCET_ADDRESSES=(
  g18qhq2fl54lszhmxeyqlvxnwjzc3xpu4nnakclp # faucet (captcha) dispensing account
  g1k28nhw04v54602jkdfrnu25gq07nyc2rehz9vl # faucet-agent dispensing account
  g15tjaykkykxa7e8nmtagm2swkphchj4j6rnukes # reserve (operator hand-sends)
)

# Vested accounts. One entry per line, in the balance-sheet vesting syntax
# (gno.land/pkg/gnoland/balance.go):
#
#   <address>=<total_coins>;vesting=<vested_coins>,<start_unix>,<end_unix>[;type=delayed]
#
# The default schedule is continuous: the vested amount unlocks linearly
# between <start_unix> and <end_unix>. `;type=delayed` makes it a cliff —
# nothing unlocks before <end_unix>, everything at once after. The vested
# coins must be <= the total, and the difference is spendable immediately.
#
# Ten shared core-team test accounts (testing-acct-1 to -10, mnemonics in the
# team password manager, keybase off-tree). 1000 GNOT each, 750 vested and 250
# spendable for fees. Every schedule completes within 7 days of GENESIS_TIME
# so each transition is observable during a two-week testnet, and the second
# week exercises the fully-vested state. The bounds are offsets from
# GENESIS_TIME so moving the launch date keeps every schedule shape.
VESTED_ACCOUNTS=(
  "g1m3efp098exf9n93l8vdh0uesftp2asr9446qh0=1000000000ugnot;vesting=750000000ugnot,$((GENESIS_TIME - 3600)),$GENESIS_TIME"                # testing-acct-1: continuous, ended at genesis: fully unlocked from block 1
  "g19em6j7376mvlf68n89k73ce3hg04zpaxf46a0e=1000000000ugnot;vesting=750000000ugnot,$GENESIS_TIME,$((GENESIS_TIME + 3600))"                # testing-acct-2: continuous, 1 hour
  "g13egct6g3d68vmrjsc5m2rrwm4767rt6wf0ytq7=1000000000ugnot;vesting=750000000ugnot,$GENESIS_TIME,$((GENESIS_TIME + 86400))"               # testing-acct-3: continuous, 1 day
  "g1uqzvq2gwakml2a6r234mw0eu0x8qck9qp42ueq=1000000000ugnot;vesting=750000000ugnot,$GENESIS_TIME,$((GENESIS_TIME + 259200))"              # testing-acct-4: continuous, 3 days
  "g14dqplajk29l0tjwr9avd5rvx6tyn7nrmkdvmxa=1000000000ugnot;vesting=750000000ugnot,$GENESIS_TIME,$((GENESIS_TIME + 604800))"              # testing-acct-5: continuous, 7 days
  "g1gdetwlyr2rw8w722whzh3jlzjmkz92pzv93jee=1000000000ugnot;vesting=750000000ugnot,$GENESIS_TIME,$((GENESIS_TIME + 3600));type=delayed"   # testing-acct-6: delayed cliff at genesis + 1 hour
  "g1dasewzx6c4fhseevfwsdfsnfy9an42kr83kzdd=1000000000ugnot;vesting=750000000ugnot,$GENESIS_TIME,$((GENESIS_TIME + 86400));type=delayed"  # testing-acct-7: delayed cliff at genesis + 1 day
  "g1pdhv586g6jl3pjml0fk5h828dad23sjlzukp4w=1000000000ugnot;vesting=750000000ugnot,$GENESIS_TIME,$((GENESIS_TIME + 604800));type=delayed" # testing-acct-8: delayed cliff at genesis + 7 days
  "g10w8l2hg7690upa4dcy6suq8yvkju7q4za0sfey=1000000000ugnot;vesting=750000000ugnot,$((GENESIS_TIME + 86400)),$((GENESIS_TIME + 345600))"  # testing-acct-9: continuous, starts 1 day after genesis, ends at 4 days
  "g17s8dlta3fjgztppcr5z7tlmkeg4ecwsfeeatag=1000000000ugnot;vesting=750000000ugnot,$((GENESIS_TIME - 604800)),$((GENESIS_TIME + 604800))" # testing-acct-10: continuous, started 7 days before genesis, half unlocked at launch, ends at 7 days
)

# =============================================================================
# Internal — everything below is glue, you shouldn't need to change it.
# =============================================================================

# Deployer key mnemonic (deterministic — used only to sign genesis-mode txs).
# Same as gnoland1/test13/topaz so the deployer address is reproducible.
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
#   <sha256>  <path-relative-to-pearl.gno.land>
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
2f686094b25ab6cb17e5ab032e98b45de8df9db8e5be1fbb76642c137290582f  work/packages.gen.txt
7717a8fc3f01fd40f6625daa6746b47fc2f023d1a664f5fea1e148ef86cab767  work/valoper-seed.jsonl
499d9fbaaea8822d873a8e6693e329c9347bc69223daf056c4f31fe28aa437dc  work/genesis_txs.jsonl

# Final artifact (moved to pearl.gno.land/ root on success)
f07b8056756dae68f15ad69c7bfa6c0da2aa2a39cbbc08bfceb9e5b454c2e0cc  genesis.json
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
# verify_checksum <path> [<manifest-key>]
#
# Computes sha256 of the file at <path>, looks up <manifest-key> (default:
# <path> relative to PEARL_DIR) in the inline CHECKSUMS_DATA heredoc,
# and one of:
#   - hash matches               → silent OK
#   - hash differs               → FAIL with expected vs got
#   - key not listed             → print computed sha256 + the line to append
verify_checksum() {
  local path="$1"
  if [ -z "${PEARL_DIR:-}" ]; then
    die "verify_checksum: PEARL_DIR not set"
  fi
  if [ ! -f "$path" ]; then
    die "verify_checksum: $path does not exist"
  fi

  local rel="${2:-${path#"$PEARL_DIR"/}}"
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
gen-genesis.sh — pearl genesis builder (single-file pipeline).

Usage:
  ./gen-genesis.sh [flags]

Flags:
  --no-install    Reuse previously built binaries in work/bin/.
  --debug         Echo the main pipeline commands before running them.
  -h, --help      Print this help and exit.

Output:
  genesis.json    Final artifact, sha256-locked against the
                  CHECKSUMS_DATA heredoc in this script.

See misc/deployments/pearl.gno.land/README.md for what the genesis
contains.
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
PEARL_DIR="$SCRIPT_DIR"
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
cleanup() { [ -n "$NODE_PID" ] && kill "$NODE_PID" 2>/dev/null && wait "$NODE_PID" 2>/dev/null || true; }
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
  local tx_json="$WORK_DIR/.txn.json"

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

  # Guard the extraction: a non-array .args yields zero --args silently,
  # and a dropped empty-string arg would shift every later positional.
  local args_len
  args_len=$(jq '.args // [] | length' "$meta")
  if [ $((${#args_array[@]} / 2)) -ne "$args_len" ]; then
    die "args mismatch in $meta: extracted $((${#args_array[@]} / 2)) of $args_len (empty or non-string args are not supported)"
  fi

  local tx_json="$WORK_DIR/.txn.json"

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

printf '\n### pearl genesis build ###\n'

# ---- Step 1: Resolve script paths and tooling

print_step_header 1 "$TOTAL_STEPS" "Resolve script paths and tooling"

GENESIS_FILE="$WORK_DIR/genesis.json"
PACKAGES_GEN_FILE="$WORK_DIR/packages.gen.txt"
GENESIS_TXS_JSONL="$WORK_DIR/genesis_txs.jsonl"
DEPLOYER_BALANCES="$WORK_DIR/deployers_balances.txt"
VALOPER_CSV="$WORK_DIR/valoper_profiles.csv"
VALOPER_SEED="$WORK_DIR/valoper-seed.jsonl"

print_substep "1.1" "PEARL_DIR=$PEARL_DIR"
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
# -test-dep also resolves test-only imports (uassert, urequire, ...).
# Packages ship on-chain with their _test.gno files (MPUserAll), but
# addpkg type-checks production files only (tests are stored and
# syntax-parsed), so a missing test dep does not fail the deploy — the
# deps are included so the shipped test files keep their imports
# resolvable on-chain.
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
# Rebuild staging from scratch — leftovers from a previous partial run
# would be addpkg'd into the genesis alongside the current set.
WORK_DIR_EXAMPLES="$WORK_DIR/examples"
rm -rf "$WORK_DIR_EXAMPLES"
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

# DEPLOYER_ADDR is load-bearing three ways (valoper fee payer, readiness
# probe, measured funding) — assert it matches the recovered key.
deployer_listed=$("$GNOKEY_BIN" list --home "$WORK_DIR_GNOKEY_HOME" 2>/dev/null || true)
case "$deployer_listed" in
*"$DEPLOYER_ADDR"*) ;;
*) die "DEPLOYER_ADDR $DEPLOYER_ADDR is not the address derived from DEPLOYER_MNEMONIC" ;;
esac

print_substep "4.5" "Generating empty genesis..."
run "$GNOGENESIS_BIN" generate -chain-id "$CHAIN_ID" -genesis-time "$GENESIS_TIME" --output-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'

print_substep "4.6" "Adding $pkg_count packages to genesis..."
echo "" | run "$GNOGENESIS_BIN" txs add packages "$WORK_DIR_EXAMPLES" -gno-home "$WORK_DIR_GNOKEY_HOME" -key-name "$DEPLOYER_KEY" --genesis-path "$GENESIS_FILE" --insecure-password-stdin 2>&1 | sed 's/^/    /'

print_substep "4.7" "Exporting txs..."
run "$GNOGENESIS_BIN" txs export "$GENESIS_TXS_JSONL" --genesis-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'

# ---- Step 5: Add the bootstrap MsgRun (transactions/base/bootstrap/)
# Seeds the sole GovDAO T1 member (aeddi) and locks AllowedDAOs. No
# transfer lock, no unrestricted-accounts proposals — pearl transfers
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

# The admin-gated call is the single security-relevant genesis tx; the
# caller the chain will trust comes from meta.json's caller_override, not
# from NAMES_ADMIN — assert they agree instead of logging an unverified
# value.
names_caller=$(jq -r '.caller_override // empty' "$NAMES_ENABLE_DIR/meta.json")
if [ "$names_caller" != "$NAMES_ADMIN" ]; then
  die "names-enable caller_override ($names_caller) != NAMES_ADMIN ($NAMES_ADMIN)"
fi

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
    printf '%s,%s,%s,pearl founding validator (%s),cloud\n' \
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

# `txs add sheets` dedupes against the genesis while the jsonl is a plain
# concat — a duplicated tx would be measured (and funded) in step 8 but
# not shipped. Reconcile the two before measuring.
genesis_tx_count=$(jq '.app_state.txs | length' "$GENESIS_FILE")
if [ "$genesis_tx_count" -ne "$tx_count" ]; then
  die "genesis holds $genesis_tx_count txs but the exported stream has $tx_count — duplicated or dropped txs"
fi

# ---- Step 8: Calculate genesis fee-payer balances
# Same approach as gnoland1's gen-genesis.sh: spin up a temp node, pre-fund
# every creator/caller address with $INITIAL_BALANCE, let the genesis txs
# burn through fees, then query remaining balances. The amount actually
# spent is what we credit each fee payer in the real genesis so their
# balance lands at zero post-genesis — the final state then holds ONLY the
# faucet balances. The fee payers are the deployer (addpkgs, bootstrap,
# valoper Register fees), the names admin (names.Enable fee), and every
# gnomod.toml [addpkg] creator address in the package set (used as that
# package's addpkg creator in place of the deployer).
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

rm -rf "$BALANCES_TMP_DIR"
mkdir -p "$BALANCES_TMP_DIR"

print_substep "8.1" "Extracting creator/caller addresses..."
# The extraction below assumes these msg types (their signer field is
# creator/caller). Any other type names its fee payer differently and
# would end up unfunded — and the chain auto-mints genesis funding for
# unfunded signers, silently shipping an unaccounted balance.
unexpected_types=$(jq -r '.tx.msg[]["@type"]' "$GENESIS_TXS_JSONL" | sort -u |
  grep -vE '^/vm\.(m_addpkg|m_run|m_call)$' || true)
if [ -n "$unexpected_types" ]; then
  die "unexpected msg types in genesis txs: $unexpected_types"
fi
grep -oE '"(creator|caller)":"[^"]*"' "$GENESIS_TXS_JSONL" |
  sed 's/"creator":"//;s/"caller":"//;s/"//g' |
  awk 'NF' |
  sort -u >"$BALANCES_TMP_CREATOR_ADDRESSES"
addr_count=$(wc -l <"$BALANCES_TMP_CREATOR_ADDRESSES" | tr -d ' ')
print_substep "8.2" "Found $addr_count unique creator/caller addresses"

# A genesis-funded address (faucet or vested) must not also be a fee
# payer: the final balance sheet keeps one entry per address (last write
# wins), so a collision would silently replace the exact-burn entry — and
# the measurement below would be wrong for that address anyway. No address
# may appear on two funded lists either. Fail fast before spinning up
# nodes.
FUNDED_ADDRESSES_FILE="$BALANCES_TMP_DIR/funded-addrs.txt"
{
  for faucet in "${FAUCET_ADDRESSES[@]}"; do
    echo "$faucet"
  done
  for vested in "${VESTED_ACCOUNTS[@]}"; do
    # Balance.Parse trims whitespace, so a padded entry would parse fine
    # downstream while its extracted address silently misses the guard —
    # reject malformed entries here instead of failing open.
    case "$vested" in
    *[[:space:]]*) die "malformed VESTED_ACCOUNTS entry (contains whitespace): '$vested'" ;;
    *=*) ;;
    *) die "malformed VESTED_ACCOUNTS entry (missing '='): '$vested'" ;;
    esac
    echo "${vested%%=*}"
  done
} >"$FUNDED_ADDRESSES_FILE"
funded_dupes=$(sort "$FUNDED_ADDRESSES_FILE" | uniq -d)
if [ -n "$funded_dupes" ]; then
  die "address(es) appearing more than once across the funded lists (faucets + vested): $funded_dupes"
fi
while IFS= read -r funded; do
  if grep -qxF -- "$funded" "$BALANCES_TMP_CREATOR_ADDRESSES"; then
    die "funded account $funded is also a genesis-tx fee payer — its exact-burn entry would be silently overwritten"
  fi
done <"$FUNDED_ADDRESSES_FILE"

# Pre-flight the vested entries: `balances add` catches syntax errors
# (Balance.Parse) and `verify` catches schedule semantics (Balance.Verify
# — verify is its only caller), so a malformed entry dies here in seconds
# instead of after two temp-node runs.
if [ "${#VESTED_ACCOUNTS[@]}" -gt 0 ]; then
  VESTED_PREFLIGHT_GENESIS="$BALANCES_TMP_DIR/vested-preflight.json"
  VESTED_PREFLIGHT_SHEET="$BALANCES_TMP_DIR/vested-preflight-sheet.txt"
  for vested in "${VESTED_ACCOUNTS[@]}"; do
    echo "$vested"
  done >"$VESTED_PREFLIGHT_SHEET"
  run "$GNOGENESIS_BIN" generate -chain-id "$CHAIN_ID" -genesis-time "$GENESIS_TIME" -output-path "$VESTED_PREFLIGHT_GENESIS" >/dev/null 2>&1
  run "$GNOGENESIS_BIN" balances add -balance-sheet "$VESTED_PREFLIGHT_SHEET" -genesis-path "$VESTED_PREFLIGHT_GENESIS" >/dev/null
  # verify refuses an empty validator set; borrow the first founding
  # validator to make the throwaway genesis verifiable.
  read -r pf_name pf_power pf_address pf_pub_key <<<"${INITIAL_VALSET[0]}"
  run "$GNOGENESIS_BIN" validator add -name "$pf_name" -power "$pf_power" -address "$pf_address" -pub-key "$pf_pub_key" -genesis-path "$VESTED_PREFLIGHT_GENESIS" >/dev/null
  run "$GNOGENESIS_BIN" verify -genesis-path "$VESTED_PREFLIGHT_GENESIS" >/dev/null
  rm -f "$VESTED_PREFLIGHT_GENESIS" "$VESTED_PREFLIGHT_SHEET"
fi

# Vested entries also ride the temp-node sheets so both measurement runs
# exercise vesting-account creation in InitChain. They are not fee payers
# (the guard above), so they cannot perturb the burn measurement, and the
# zero-verify loops iterate fee payers only.
append_vested_to_sheet() {
  local vested
  for vested in "${VESTED_ACCOUNTS[@]}"; do
    echo "$vested" >>"$BALANCES_TMP_FILE"
  done
}

print_substep "8.3" "Generating over-provisioned balances..."
while IFS= read -r addr; do
  echo "${addr}=${INITIAL_BALANCE}ugnot" >>"$BALANCES_TMP_FILE"
done <"$BALANCES_TMP_CREATOR_ADDRESSES"
append_vested_to_sheet

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

  # The measured burns are dominated by storage deposits priced by the
  # chain's vm/auth params. This genesis is regenerated rather than copied
  # (a future GENESIS_TIME would stall the temp node), so assert its
  # fee-governing params equal the genesis that ships — otherwise every
  # measured amount is wrong by the parameter ratio and run 2 would agree.
  if [ "$(jq -cS '[.app_state.vm.params, .app_state.auth.params]' "$BALANCES_TMP_GENESIS")" != \
    "$(jq -cS '[.app_state.vm.params, .app_state.auth.params]' "$GENESIS_FILE")" ]; then
    die "temp-node genesis fee params differ from the shipping genesis — measurement would be invalid"
  fi
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

  # Readiness = the deployer account visible in committed state, not just
  # an answering RPC (see account_in_state). Balance reads are only
  # trustworthy from that point on.
  local elapsed=0
  while [ "$elapsed" -lt "$NODE_TIMEOUT" ]; do
    if ! kill -0 "$NODE_PID" 2>/dev/null; then
      echo "ERROR: Node stopped unexpectedly. Last log lines:" >&2
      tail -20 "$BALANCES_TMP_GNOLAND_LOG" >&2
      exit 1
    fi
    if account_in_state "$DEPLOYER_ADDR"; then
      printf "  Node ready (%ss)\n" "$elapsed"
      return
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  kill "$NODE_PID" 2>/dev/null || true
  echo "ERROR: Node did not start within ${NODE_TIMEOUT}s." >&2
  printf 'last probe response: %s\n' \
    "$("$GNOKEY_BIN" query -remote "$NODE_RPC_ADDR" "auth/accounts/$DEPLOYER_ADDR" 2>&1 || true)" >&2
  echo "Last log lines:" >&2
  tail -20 "$BALANCES_TMP_GNOLAND_LOG" >&2
  exit 1
}

stop_temp_node() {
  kill "$NODE_PID" 2>/dev/null || true
  wait "$NODE_PID" 2>/dev/null || true
  NODE_PID=""
}

# account_in_state ADDR → success if ADDR's account exists in committed
# state. Queries read the last committed block: while InitChain is still
# executing (or block 1 is not yet committed) the store reads empty and
# auth/accounts answers `data: null` with exit 0, so a bare exit-status
# probe cannot tell "node up" from "state committed" — the account JSON
# can. Genesis state (balance credits included) commits atomically with
# block 1, so any funded address proves the whole genesis is readable.
# Matched with a bash pattern, not a pipe to grep: grep -q closing the
# pipe early can SIGPIPE the producer, which pipefail would report as a
# failed probe despite a match.
account_in_state() {
  local addr="$1" out
  out=$("$GNOKEY_BIN" query -remote "$NODE_RPC_ADDR" "auth/accounts/$addr" 2>/dev/null || true)
  case "$out" in
  *"\"address\": \"$addr\""*) return 0 ;;
  *) return 1 ;;
  esac
}

# query_balance ADDR → echoes ugnot amount.
# An empty `data:` response is how a zero balance is encoded (empty
# coins) — but it is also what an unreadable store returns. A zero is
# only echoed after account_in_state proves the address exists in
# committed state; otherwise the build dies.
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
      local payload
      payload=$(echo "$out" | sed -n 's/^data: //p' | head -1)
      if [ "$payload" = '""' ]; then
        # Empty coins encode a zero balance — but an unreadable store
        # answers identically. Only trust the zero if the account
        # provably exists in committed state.
        account_in_state "$addr" ||
          die "balance for $addr reads zero but its account is not in committed state"
        echo 0
        return
      fi
      # Exactly one ugnot coin and nothing else: a second denomination on
      # a fee payer must abort the measurement, not parse as 0.
      local r
      r=$(echo "$payload" | sed -n 's/^"\([0-9][0-9]*\)ugnot"$/\1/p')
      [ -n "$r" ] || die "unparseable balance for $addr: $payload (expected a single ugnot coin or empty)"
      echo "$r"
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
  # A fee payer funded with the float and charged at least one fee must
  # land strictly between 0 and INITIAL_BALANCE.
  if [ "$remaining" -eq 0 ] || [ "$remaining" -ge "$INITIAL_BALANCE" ]; then
    die "fee payer $addr reads $remaining ugnot remaining in the measure run (float: $INITIAL_BALANCE) — node state not readable or fees not charged"
  fi
  final=$((INITIAL_BALANCE - remaining))
  printf "    %s = %s ugnot\n" "$addr" "$final"
  echo "${addr}=${final}ugnot" >>"$BALANCES_TMP_FILE"
done <"$BALANCES_TMP_CREATOR_ADDRESSES"
append_vested_to_sheet
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
# The temp sheet also carries the vested entries (so the measurement runs
# exercise their creation); keep only the measured fee-payer lines here —
# step 9 appends the vested entries itself.
grep -vF -- ';vesting=' "$BALANCES_TMP_FILE" >"$DEPLOYER_BALANCES"

# ---- Step 9: Add validators + balances, verify, move into place

print_step_header 9 "$TOTAL_STEPS" "Add validators + balances, verify genesis"

print_substep "9.1" "Adding the initial validator set..."
for validator in "${INITIAL_VALSET[@]}"; do
  read -r name power address pub_key <<<"$validator"
  printf "    %s (power=%s, %s)\n" "$name" "$power" "$address"
  run "$GNOGENESIS_BIN" validator add -name "$name" -power "$power" -address "$address" -pub-key "$pub_key" --genesis-path "$GENESIS_FILE"
done

# Fee payers (exact-burn, land at zero) + the faucets + the vested
# accounts, one sheet. One entry per address (last write wins), so an
# address on two lists would lose one of its entries — the overlap guard
# in step 8 rules that out. Vested entries already carry the full
# balance-sheet syntax (amount + vesting schedule) and are appended
# verbatim.
FULL_BALANCES_FILE="$WORK_DIR/balances.txt"
cp "$DEPLOYER_BALANCES" "$FULL_BALANCES_FILE"
for addr in "${FAUCET_ADDRESSES[@]}"; do
  echo "${addr}=${FAUCET_BALANCE}ugnot" >>"$FULL_BALANCES_FILE"
done
for vested in "${VESTED_ACCOUNTS[@]}"; do
  echo "$vested" >>"$FULL_BALANCES_FILE"
done
balance_count=$(wc -l <"$FULL_BALANCES_FILE" | tr -d ' ')
print_substep "9.2" "Adding $balance_count balances (fee payers + ${#FAUCET_ADDRESSES[@]} faucets + ${#VESTED_ACCOUNTS[@]} vested)..."
run "$GNOGENESIS_BIN" balances add -balance-sheet "$FULL_BALANCES_FILE" --genesis-path "$GENESIS_FILE" >/dev/null

print_substep "9.3" "Running gnogenesis verify..."
# -skip-signature-check: the names.Enable tx carries a post-sign caller
# patch and the valoper Register txs carry placeholder signatures, so
# per-tx signature verification cannot pass by design (nodes accept both
# under --skip-genesis-sig-verification). The remaining scope is shallow —
# amino decode, GenesisDoc/params sanity, stateless per-tx and per-balance
# format checks. Execution and fee-payer funding are proven by step 8's
# temp-node runs, not by this step.
run "$GNOGENESIS_BIN" verify -genesis-path "$GENESIS_FILE" -skip-signature-check

# Verify before moving: a mismatch must not clobber the previously-good
# (gitignored, so invisible to git status) genesis.json at the root.
verify_checksum "$GENESIS_FILE" genesis.json
print_substep "9.4" "Moving $GENESIS_FILE -> $FINAL_GENESIS"
mv "$GENESIS_FILE" "$FINAL_GENESIS"

# ---- Summary

PIPELINE_END_TS=$(date +%s)
PIPELINE_DURATION=$((PIPELINE_END_TS - PIPELINE_START_TS))
FINAL_SHA=$(sha256_of "$FINAL_GENESIS")
FINAL_BYTES=$(file_size "$FINAL_GENESIS")

printf '\n### pearl build complete: genesis.json (%s, sha256=%s) ###\n' \
  "$(format_size "$FINAL_BYTES")" "$FINAL_SHA"
printf '    total pipeline time: %s\n' "$(format_duration "$PIPELINE_DURATION")"
