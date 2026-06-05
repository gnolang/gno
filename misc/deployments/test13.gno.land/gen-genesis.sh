#!/usr/bin/env bash
# gen-genesis.sh — test-13 hardfork genesis builder (single-file pipeline).
#
# Replaces the former build.sh + phase-1-build-genesis.sh +
# phase-2-apply-replay.sh + lib/common.sh + CHECKSUMS.txt with one
# self-contained script. The sha256 manifest is embedded as the
# CHECKSUMS_DATA heredoc below.
#
# Two conceptual phases run in sequence:
#
#   Phase 1: build the BASE genesis (no historical replay yet). 9 steps.
#            Resolves packages, calculates deployer balances, adds the
#            test-13 validator set + faucets, and emits the valoper-seed
#            migration jsonl.
#
#   Phase 2: apply gnoland1's historical txs + patches + migration txs to
#            produce the final genesis. 5 steps. Builds the T1 rotation +
#            names-enable migration txs, runs `gnogenesis fork generate`,
#            verifies via in-process audit.
#
# Phase 1 deltas from gnoland1's gen-genesis.sh that make this base
# "hardfork-ready":
#
#   1. FILTERED_PACKAGES adds the four packages that gnoland1 doesn't ship
#      (p/onbloc/{uint256,int256,json}, r/sys/validators/v3) and one realm
#      (r/demo/defi/grc20reg) we want available post-fork.
#   2. INITIAL_VALSET is the test-13 valset rather than gnoland1's launch
#      7. Power 10 each — chosen for visibility in tooling that displays
#      the ratio of votes; consensus only cares about relative weight.
#   3. transactions/base/bootstrap/govdao_prop1_test13.gno replaces gnoland1's
#      govdao_prop1.gno. It drops the v2 valset seed (we set the consensus
#      valset directly via GenesisDoc.Validators), adds 10 faucet
#      addresses to the ProposeAddUnrestrictedAcctsRequest call so we can
#      transact under the restricted-denom regime without manual unrestrict
#      txs, and wires the packages/rotate realm (delta 4)
#      into AllowedDAOs at lock time.
#   4. packages/rotate/rotate.gno is a single-use realm
#      addpkg'd alongside the filtered set. A phase-2 MsgCall to its
#      Rotate() function swaps the sole T1 from manfred (gnoland1 inherit)
#      to the test-13 T1, then Rotate self-ejects from AllowedDAOs.
#   5. INITIAL_VALSET_OPERATORS pairs each founding validator with a
#      distinct operator address; phase-1 step 8 runs `gnogenesis fork
#      valoper-seed` to emit a .jsonl of genesis-mode valopers.Register
#      MsgCalls. Without this, the founding validators boot as orphans
#      (no v3 operator-keyed management plane) and gnoland's InitChainer
#      AssertGenesisValopersConsistent fires (#5701 makes this loud,
#      #5702 catches it at verify time).
#
# Phase 2 audit step:
#   `gnogenesis fork test --verbose --skip-failing-genesis-txs` replays the
#   assembled genesis in-process; an in-memory gnoland node executes every
#   genesis tx (base addpkg + base MsgRun + historical + patched + migration)
#   and any tx that panics/errors is logged. We require 0 failures; any
#   non-zero count aborts the build with a top-10-by-frequency print. This
#   is the only gate that confirms the assembled genesis actually boots
#   and runs end-to-end — gnogenesis verify only checks structural
#   integrity (JSON validity, balance sums).
#
# Output:
#   work/phase-1/base-genesis.json    base genesis consumed by phase-2
#   work/phase-1/valoper-seed.jsonl   valoper Register migration txs
#   work/phase-2/genesis.json         intermediate phase-2 output
#   work/phase-2/t1-rotation.jsonl    T1 + names migration txs
#   work/phase-2/fork-test.log        full audit log
#   genesis.json                      final hardfork artifact (192 MB-ish)
#
# Usage:
#   ./gen-genesis.sh                                 # full build (both phases)
#   ./gen-genesis.sh --phase 1                       # phase 1 only
#   ./gen-genesis.sh --phase 2                       # phase 2 only (requires phase-1 artifacts)
#   ./gen-genesis.sh --debug                         # show every command being run
#   ./gen-genesis.sh --no-install                    # reuse previously built binaries
#   ./gen-genesis.sh --skip-audit                    # phase-2 only: skip audit step
#   ./gen-genesis.sh --source-txs-jsonl-file PATH    # phase-2 only: use cached jsonl
#   ./gen-genesis.sh --source-txs-rpc URLS           # phase-2 only: multi-endpoint RPC fetch
#   ./gen-genesis.sh --source-txs-data-dir PATH      # phase-2 only: read from halted gnoland data dir
#
# Cross-platform: bash 3.2 minimum (macOS default), no GNU-only features.
# Every external tool routed through a dispatcher with fallbacks (curl→wget,
# shasum→sha256sum, gunzip→gzip).

set -eo pipefail

# =============================================================================
# Launch parameters — review before each genesis generation.
# =============================================================================

CHAIN_ID=test-13
ORIGINAL_CHAIN_ID=gnoland1
GENESIS_TIME=1773651600 # Monday, March 16th 2026 10:00 GMT+0100 (CET) — same as gnoland1

# Source-chain halt height. Passed to gnogenesis fork generate as
# --halt-height and enforced on the fetch/read side for every txs source.
# In --source-txs-jsonl-file mode, the script also cross-checks the cached
# archive's max BlockHeight against this constant to fail fast on a stale
# cache. The resulting chain starts at InitialHeight = HALT_HEIGHT + 1.
HALT_HEIGHT=1485629

# Packages to include in genesis (resolved with transitive dependencies).
# Use "..." suffix to match all sub-packages.
#
# First seven lines mirror gnoland1's gen-genesis.sh FILTERED_PACKAGES. The
# last block is test-13's additions:
#   - p/onbloc/{uint256,int256,json}: used by realms we want available on
#     test-13 but absent from gnoland1's source genesis (uint256 is a
#     transitive dep of int256). gnogenesis txs add packages resolves the
#     full dep graph from these entries.
#   - r/sys/validators/v3: PR #5485's valset realm. Master's EndBlocker
#     reads valset state from this realm's params; without it on chain,
#     post-genesis valset changes can't happen. Mainnet gnoland1 doesn't
#     deploy it, so we addpkg it here.
#   - r/demo/defi/grc20reg: GRC20 token registry; not in gnoland1's filter.
FILTERED_PACKAGES=(
  ./gno.land/r/sys/...
  ./gno.land/r/gov/...
  ./gno.land/r/gnoland/blog/...
  ./gno.land/r/gnoland/wugnot/...
  ./gno.land/r/gnoland/coins/...
  ./gno.land/r/gnoland/boards2/...
  ./gno.land/r/gnops/valopers/...
  # test-13 additions:
  ./gno.land/p/onbloc/uint256
  ./gno.land/p/onbloc/int256
  ./gno.land/p/onbloc/json
  ./gno.land/r/sys/validators/v3
  ./gno.land/r/demo/defi/grc20reg
)

# Initial test-13 validator set. Format: "name power address pub_key".
# Power 10 each (cosmetic — consensus is about ratios, not absolutes).
INITIAL_VALSET=(
  "gfanton-1 10 g15sysd4jcpsw7t0n4ffe2hn8ndfup2ae2vwpves gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqedky7xlkj9vzgtxvyh8um8vzmp3caahmvvsq5z4c9qhz0p8emsjcnhrq0"
  "aeddi-1 10 g1ahsucf2e2g95n7xufes8mqtuucw5829t044s22 gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpfe66wa78xkx5rjgmuuekaldwk8n7xcv270dfthj8x674fvsmrr46xtcfw"
)

# Operator address for each INITIAL_VALSET entry (same index). MUST be
# distinct from the signing address — `gnogenesis fork valoper-seed`
# rejects operator==signing_addr to keep signing-key compromise from
# collapsing into operator-slot compromise (see valoper_seed.go).
#
# Without these profiles seeded, gnoland's InitChainer-side
# AssertGenesisValopersConsistent fires (#5701 makes this a hard panic)
# and `gnogenesis verify` fails the new valoper-coverage check (#5702).
# v3's operator-keyed proposal flow also has no way to manage the
# founding set without these, leaving orphan validators.
#
# TODO(test-13): replace placeholders with addresses the validator
# owner actually controls before redeploying.
INITIAL_VALSET_OPERATORS=(
  "g19rl4cm2hmr8afy4kldpxz3fka4jguq0a0u3773" # node-1 operator (placeholder — BIP-39 test mnemonic, not for production)
  "g1avgyh77ycn997ja45q5q8ss8y9mr424jnx7s59" # node-2 operator (placeholder — BIP-39 test mnemonic, not for production)
)

# Faucet balances. Each gets $FAUCET_BALANCE ugnot at genesis. Addresses
# are pasted from `gnokey list` output of an off-tree keybase (mnemonics
# are NOT in this repo). The faucet addresses MUST also appear in
# transactions/base/bootstrap/govdao_prop1_test13.gno's ProposeAddUnrestrictedAcctsRequest
# call so they can transfer ugnot under the locked-bank regime; the two
# lists are kept in sync by hand for now (10 entries, low maintenance churn).
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

# Airdrop balances — reused as-is from gnoland1. The fork promise is to
# preserve gnoland1's balance state, so the airdrop snapshot lands here
# verbatim.
BALANCES_GZ_URL="https://github.com/gnolang/independence-day/raw/9dec38a4a72c9e84db7e78ae010370de250f2d64/mkgenesis/balances.txt.gz"

# Default RPC endpoint list for --source-txs-rpc. Comma-separated for
# #5693's multi-endpoint parallel fetcher. Used when no --source-txs-*
# flag is passed.
DEFAULT_TXS_RPC_ENDPOINTS="http://51.159.14.234:26657,http://163.172.33.181:26657,https://rpc.gnoland1.moul.p2p.team,https://rpc.gnoland1-aeddi-1.gnoland.network,https://rpc.gnoland1-gfanton-1.gnoland.network"

# =============================================================================
# Internal — everything below is glue, you shouldn't need to change it.
# =============================================================================

# Deployer key mnemonic (deterministic — used only to sign genesis-mode txs).
# Same as gnoland1 so the deployer address is reproducible across both chains.
DEPLOYER_MNEMONIC="anchor hurt name seed oak spread anchor filter lesson shaft wasp home improve text behind toe segment lamp turn marriage female royal twice wealth"
DEPLOYER_KEY=GenesisDeployer

# r/sys/names admin: hardcoded in examples/gno.land/r/sys/names/verifier.gno
# (the gnoland1 GovDAO T1 multisig). names.Enable's admin check reads
# runtime.PreviousRealm().Address(); under --skip-genesis-sig-verification,
# the chain trusts the MsgCall.Caller field as the EOA, so jq-patching
# caller to this address makes Enable's gate pass.
NAMES_ADMIN=g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh

# ---- Locked sha256 hashes (formerly CHECKSUMS.txt).
#
# Format (matches `shasum -a 256` / `sha256sum` output exactly):
#   <sha256>  <path-relative-to-test13.gno.land>
#
# Two spaces between hash and path. Blank lines and `#`-prefixed lines are
# ignored. Each phase block calls `verify_checksum <path>` after producing
# or downloading an artifact:
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
# Phase 1 artifacts
db510a244428bf22efb9a2b1e1c5c88bf4cb8d020714868066a389300475ab02  work/phase-1/airdrop_balances.txt.gz
b1e628ba3172ed2801360ef8270746c4d301350bf3023977e728ef10f65c739a  work/phase-1/packages.gen.txt
11ee1bb4d93c11c29140e2dcc705b34213482196181c6eabef1f2bd0f5cc77c7  work/phase-1/genesis_txs.jsonl
b2932463190d60bab2763cd5c8dadf010556da226f73426713cec1d79d344e67  work/phase-1/base-genesis.json
197a642b82b6176ad8af2879bb8b29e781dd06ce499d3dc03ea3e4fbc5141460  work/phase-1/valoper-seed.jsonl

# Phase 2 artifacts
d0eb55f3c7221771956abbc2b8cd49fae08baebe571194143fff3f48c42a6df7  work/phase-2/txs.jsonl
46bc04027359f61d90aef274c84e0b3ee25b78d49cc691842e16a7e0219ff584  work/phase-2/t1-rotation.jsonl
2ea3bbe7dd874c5a8003446eb69aca668333a5cf71f50f1d6f2191c2a9dd3930  work/phase-2/genesis.json

# Final artifact (moved to test13.gno.land/ root on phase-2 success)
2ea3bbe7dd874c5a8003446eb69aca668333a5cf71f50f1d6f2191c2a9dd3930  genesis.json
EOF
)

# =============================================================================
# Helper functions (formerly lib/common.sh).
# =============================================================================

# ---- Fatal error reporter

die() {
  printf 'ERROR: %s\n' "$*" >&2
  exit 1
}

# ---- Tool dispatchers
# Inline `command -v X` checks in the phase blocks are forbidden — use these.

# download_url <url> <dest_path>
# Tries curl, then wget. Both with sane flags. Errors if neither is available.
download_url() {
  local url="$1"
  local dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$dest"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$dest" "$url"
  else
    die "neither curl nor wget is installed (need one of them to download $url)"
  fi
}

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

# gunzip_to_stdout <gz_path>
# Prints the decompressed content to stdout. Tries gunzip -c, falls back to
# gzip -dc. Errors if neither is available.
gunzip_to_stdout() {
  local gz="$1"
  if command -v gunzip >/dev/null 2>&1; then
    gunzip -c "$gz"
  elif command -v gzip >/dev/null 2>&1; then
    gzip -dc "$gz"
  else
    die "neither gunzip nor gzip is installed (need one of them to decompress $gz)"
  fi
}

# ---- Tool preflight
# require_tools <tool>...
# Probes every named tool; if any are missing, prints the full list with
# install hints (apt + brew) and exits. Used at the top of each phase
# block so the operator sees all missing tools at once instead of failing
# on the first occurrence mid-run.
#
# Recognised tools (each with a known install hint):
#   curl, wget   — at least one needed (download_url tries both)
#   shasum, sha256sum — at least one needed (sha256_of tries both)
#   gunzip, gzip — at least one needed (gunzip_to_stdout tries both)
#   jq           — required (no fallback)
#   awk, sed, grep, sort, tr, mv, cp, ls, find, wc, head, tail, cut — POSIX core
#   tar          — used by gnogenesis dependencies
#   go           — required to build the gno binaries
#   python3      — used by pick_free_port helper in phase-1
#
# Logical groups (at-least-one): pass any of the alternatives — the function
# treats them as alternatives only when ALL are missing. Otherwise each name
# is checked independently.
require_tools() {
  local missing=""
  local tool
  for tool in "$@"; do
    case "$tool" in
    "curl|wget")
      if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        missing="$missing curl|wget"
      fi
      ;;
    "shasum|sha256sum")
      if ! command -v shasum >/dev/null 2>&1 && ! command -v sha256sum >/dev/null 2>&1; then
        missing="$missing shasum|sha256sum"
      fi
      ;;
    "gunzip|gzip")
      if ! command -v gunzip >/dev/null 2>&1 && ! command -v gzip >/dev/null 2>&1; then
        missing="$missing gunzip|gzip"
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
    "curl|wget")
      printf '      install:  brew install curl   |   apt-get install -y curl\n' >&2
      ;;
    "shasum|sha256sum")
      printf '      install:  brew install coreutils   |   apt-get install -y coreutils\n' >&2
      ;;
    "gunzip|gzip")
      printf '      install:  brew install gzip   |   apt-get install -y gzip\n' >&2
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
    awk | sed | grep | sort | tr | mv | cp | ls | find | wc | head | tail | cut | tar)
      printf '      install:  comes with any POSIX userland (coreutils + findutils + tar)\n' >&2
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
# TEST13_DIR) in the inline CHECKSUMS_DATA heredoc, and one of:
#   - hash matches               → silent OK
#   - hash differs               → FAIL with expected vs got
#   - path not listed            → print computed sha256 + the line to append
verify_checksum() {
  local path="$1"
  if [ -z "${TEST13_DIR:-}" ]; then
    die "verify_checksum: TEST13_DIR not set"
  fi
  if [ ! -f "$path" ]; then
    die "verify_checksum: $path does not exist"
  fi

  local rel="${path#"$TEST13_DIR"/}"
  local got
  got=$(sha256_of "$path")

  # Find the line whose second field equals $rel inside the inline manifest.
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
# print_phase_header <phase> <step> <total> <title>
#   prints e.g. `=== Phase 1 / Step 3 / 9: Resolve script paths and tooling ===`
print_phase_header() {
  local phase="$1"
  local step="$2"
  local total="$3"
  local title="$4"
  printf '\n=== Phase %s / Step %s of %s: %s ===\n' \
    "$phase" "$step" "$total" "$title"
}

# print_substep <code> <text>
#   prints e.g. `  [1.3.1] Checking cache: work/phase-1/airdrop_balances.txt.gz`
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
# Prints "245 MB", "4 KB", "789 B". Decimal units (1000-based) — matches
# du output on macOS with BLOCKSIZE=1000. Avoids parsing `du -h`, which
# differs across platforms.
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

PHASE=all
DEBUG=false
NO_INSTALL=false
SKIP_AUDIT=false
SOURCE_TXS_JSONL_FILE=""
SOURCE_TXS_RPC=""
SOURCE_TXS_DATA_DIR=""

require_arg() {
  if [ "$#" -lt 2 ]; then
    echo "ERROR: $1 requires a value" >&2
    exit 1
  fi
}

print_usage() {
  cat <<'EOF'
gen-genesis.sh — test-13 hardfork genesis builder (single-file pipeline).

Usage:
  ./gen-genesis.sh [flags]

Flags:
  --phase 1|2|all                 Which phase to run. Default: all.
                                  1 = build base genesis (no historical replay).
                                  2 = apply historical txs + patches + migrations
                                      (requires phase-1 artifacts in work/phase-1/).
  --no-install                    Reuse previously built binaries in work/phase-1/bin/.
  --debug                         Echo every external command before running it.
  --skip-audit                    Phase 2 only: skip the in-process fork-test replay.

  --source-txs-jsonl-file PATH    Phase 2 only: read historical txs from a cached
                                  amino-jsonl file (e.g. work/phase-2/txs.jsonl).
  --source-txs-rpc URLS           Phase 2 only: fetch historical txs from comma-
                                  separated RPC endpoints (parallel multi-endpoint
                                  fetcher). Default if no --source-txs-* flag set.
  --source-txs-data-dir PATH      Phase 2 only: read historical txs offline from a
                                  halted gnoland data dir (PebbleDB).

  -h, --help                      Print this help and exit.

Output:
  genesis.json                    Final hardfork artifact, sha256-locked against
                                  the CHECKSUMS_DATA heredoc in this script.

See misc/deployments/test13.gno.land/README.md for the architecture, transaction
format spec, per-patch rationale, and validator workflow.
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
  -h | --help)
    print_usage
    exit 0
    ;;
  --phase)
    require_arg "$@"
    PHASE="$2"
    shift 2
    ;;
  --phase=*)
    PHASE="${1#*=}"
    shift
    ;;
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
    echo "ERROR: Unknown argument: $1" >&2
    echo "Run with --help for usage." >&2
    exit 1
    ;;
  esac
done

case "$PHASE" in
1 | 2 | all) ;;
*)
  die "--phase must be one of: 1, 2, all (got: $PHASE)"
  ;;
esac

# Resolve txs source — default to multi-RPC fetch if no flag given.
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

# =============================================================================
# Shared paths + cleanup trap.
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST13_DIR="$SCRIPT_DIR"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
EXAMPLES_DIR="$REPO_ROOT/examples"

PHASE1_DIR="$SCRIPT_DIR/work/phase-1"
PHASE2_DIR="$SCRIPT_DIR/work/phase-2"
WORK_DIR_BIN="$PHASE1_DIR/bin"
WORK_DIR_GNOKEY_HOME="$PHASE1_DIR/gnokey-home"

GNO_CMD="$REPO_ROOT/gnovm/cmd/gno"
GNOKEY_CMD="$REPO_ROOT/gno.land/cmd/gnokey"
GNOLAND_CMD="$REPO_ROOT/gno.land/cmd/gnoland"
GNOGENESIS_CMD="$REPO_ROOT/contribs/gnogenesis"
GNO_BIN="$WORK_DIR_BIN/gno"
GNOKEY_BIN="$WORK_DIR_BIN/gnokey"
GNOLAND_BIN="$WORK_DIR_BIN/gnoland"
GNOGENESIS_BIN="$WORK_DIR_BIN/gnogenesis"

FINAL_GENESIS="$SCRIPT_DIR/genesis.json"

# Clean up temp node on exit (phase-1 only starts it; trap is a no-op when
# NODE_PID is unset, so it's safe to install at script scope).
NODE_PID=""
cleanup() { [ -n "$NODE_PID" ] && kill "$NODE_PID" 2>/dev/null || true; }
trap cleanup EXIT

# =============================================================================
# Transaction loader — converts transactions/<...>/<txdir>/{meta.json + optional
# body.gno or pkg/} into one AnnotatedTx jsonl line appended to <outfile>.
#
# Dispatches on meta.json's "kind" field:
#   MsgRun       — base/migration MsgRun. Signs via gnokey maketx run + sign.
#   MsgCall      — base/migration MsgCall. Signs via gnokey maketx call + sign;
#                  optionally jq-patches msg[0].caller to caller_override post-sign
#                  (for genesis-mode calls that need admin caller without holding
#                  the admin key — same trick emit_migration_msgcall used to do).
#   Patched      — historical-tx override. meta.json contains a full AnnotatedTx
#                  with body slot stripped. If body_file is set, the file's
#                  content is inlined into tx.msg[0].package.files[0].body. If
#                  pkg_dir is set, ALL files under pkg/ become tx.msg[0].package.files.
#                  Caller-swap patches have neither and are emitted as-is.
#
# All emitted jsonl lines are in AnnotatedTx shape {tx, metadata, reason} —
# strip the reason field (jq -c 'del(.reason)') before passing to consumers
# that don't speak AnnotatedTx (e.g. `gnogenesis txs add sheets`).
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
  Patched) _txn_patched "$dir" "$outfile" ;;
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

  local reason caller_key caller_override pkgpath func gas_wanted gas_fee acct_num seq bh
  reason=$(jq -r '.reason' "$meta")
  caller_key=$(jq -r '.caller_key' "$meta")
  caller_override=$(jq -r '.caller_override // empty' "$meta")
  pkgpath=$(jq -r '.pkgpath' "$meta")
  func=$(jq -r '.func' "$meta")
  gas_wanted=$(jq -r '.gas_wanted' "$meta")
  gas_fee=$(jq -r '.gas_fee' "$meta")
  acct_num=$(jq -r '.account_number' "$meta")
  seq=$(jq -r '.sequence' "$meta")
  bh=$(jq -r '.block_height // "0"' "$meta")

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
    jq -c --arg c "$caller_override" --arg r "$reason" --arg bh "$bh" \
      '.msg[0].caller = $c | {tx: ., metadata: {block_height: $bh}, reason: $r}' \
      "$tx_json" >>"$outfile"
  else
    jq -c --arg r "$reason" --arg bh "$bh" \
      '{tx: ., metadata: {block_height: $bh}, reason: $r}' \
      "$tx_json" >>"$outfile"
  fi
  rm -f "$tx_json"
}

_txn_patched() {
  local dir="$1" outfile="$2"
  local meta="$dir/meta.json"

  local body_file pkg_dir
  body_file=$(jq -r '.body_file // empty' "$meta")
  pkg_dir=$(jq -r '.pkg_dir // empty' "$meta")

  if [ -n "$body_file" ]; then
    # Single-file body rewrite: inline body_file content into files[0].body.
    jq -c --rawfile body "$dir/$body_file" '
      del(.kind, .body_file) |
      .tx.msg[0].package.files[0].body = $body
    ' "$meta" >>"$outfile"
  elif [ -n "$pkg_dir" ]; then
    # Multi-file: build files array from pkg/ contents (name + body per file).
    local files_json
    files_json=$(
      for f in "$dir/$pkg_dir"/*; do
        jq -n --arg n "$(basename "$f")" --rawfile b "$f" '{name: $n, body: $b}'
      done | jq -s '.'
    )
    jq -c --argjson files "$files_json" '
      del(.kind, .pkg_dir) |
      .tx.msg[0].package.files = $files
    ' "$meta" >>"$outfile"
  else
    # No body changes (caller-swap, MsgCall) — emit as-is, just drop kind.
    jq -c 'del(.kind)' "$meta" >>"$outfile"
  fi
}

# =============================================================================
# Phase 1 — build the BASE genesis (no historical replay).
# =============================================================================

run_phase_1() {
  local PHASE_START_TS
  PHASE_START_TS=$(date +%s)
  local TOTAL_STEPS=9

  printf '\n### Phase 1 — Build BASE genesis ###\n'

  # ---- Step 1: Resolve script paths and tooling

  print_phase_header 1 1 "$TOTAL_STEPS" "Resolve script paths and tooling"

  local WORK_DIR="$PHASE1_DIR"
  local GENESIS_FILE="$WORK_DIR/base-genesis.json"
  local AIRDROP_BALANCES_GZ="$WORK_DIR/airdrop_balances.txt.gz"
  local AIRDROP_BALANCES_TXT="$WORK_DIR/airdrop_balances.txt"
  local PACKAGES_GEN_FILE="$WORK_DIR/packages.gen.txt"
  local WORK_DIR_GENESIS="$WORK_DIR/genesis.json"
  local WORK_DIR_GENESIS_TXS="$WORK_DIR/genesis_txs.jsonl"
  local WORK_DIR_DEPLOYER_BALANCES="$WORK_DIR/deployers_balances.txt"
  local WORK_DIR_VALOPER_CSV="$WORK_DIR/valoper_profiles.csv"
  local WORK_DIR_VALOPER_SEED="$WORK_DIR/valoper-seed.jsonl"

  print_substep "1.1.1" "TEST13_DIR=$TEST13_DIR"
  print_substep "1.1.2" "REPO_ROOT=$REPO_ROOT"
  print_substep "1.1.3" "WORK_DIR=$WORK_DIR"

  # ---- Step 2: Verify required tools

  print_phase_header 1 2 "$TOTAL_STEPS" "Verify required tools"

  require_tools \
    "curl|wget" \
    "shasum|sha256sum" \
    "gunzip|gzip" \
    go jq python3 \
    awk sed grep sort tr mv cp ls find wc head tail cut

  print_substep "1.2.1" "All required tools present"

  # Prepare work dir; preserve bin/ when --no-install.
  if [ "$NO_INSTALL" = true ]; then
    mkdir -p "$WORK_DIR"
    find "$WORK_DIR" -mindepth 1 -maxdepth 1 ! -name bin -exec rm -rf {} + 2>/dev/null || true
  else
    rm -rf "$WORK_DIR"
  fi
  mkdir -p "$WORK_DIR_BIN"

  # ---- Step 3: Build binaries from source

  print_phase_header 1 3 "$TOTAL_STEPS" "Build binaries from source"

  if [ "$NO_INSTALL" = true ]; then
    print_substep "1.3.1" "--no-install — reusing $WORK_DIR_BIN"
    local bin
    for bin in "$GNO_BIN" "$GNOKEY_BIN" "$GNOLAND_BIN" "$GNOGENESIS_BIN"; do
      if [ ! -x "$bin" ]; then
        die "--no-install but $bin not found. Run without --no-install first."
      fi
    done
  else
    print_substep "1.3.1" "Building gno..."
    run go build -C "$GNO_CMD" -o "$GNO_BIN" .
    print_substep "1.3.2" "Building gnokey..."
    run go build -C "$GNOKEY_CMD" -o "$GNOKEY_BIN" .
    print_substep "1.3.3" "Building gnoland..."
    run go build -C "$GNOLAND_CMD" -o "$GNOLAND_BIN" .
    print_substep "1.3.4" "Building gnogenesis..."
    run go build -C "$GNOGENESIS_CMD" -o "$GNOGENESIS_BIN" .
  fi

  # ---- Step 4: Generate filtered examples genesis txs

  print_phase_header 1 4 "$TOTAL_STEPS" "Generate filtered examples genesis txs"

  print_substep "1.4.1" "Resolving dependencies..."
  local pkg_dirs pkg_count
  pkg_dirs=$(cd "$EXAMPLES_DIR" && "$GNO_BIN" tool deplist -test-dep "${FILTERED_PACKAGES[@]}")
  pkg_count=$(echo "$pkg_dirs" | wc -l | tr -d ' ')
  print_substep "1.4.2" "Resolved $pkg_count packages in topological order"

  # Save resolved package list (used for audit + tracked by CHECKSUMS).
  {
    echo "# Generated by gen-genesis.sh — do not edit."
    # shellcheck disable=SC2001 # path contains slashes; `|` as sed delimiter is cleaner than ${//} escaping
    echo "$pkg_dirs" | sed "s|$EXAMPLES_DIR/||g"
  } >"$PACKAGES_GEN_FILE"
  verify_checksum "$PACKAGES_GEN_FILE"

  print_substep "1.4.3" "Copying packages to staging..."
  local WORK_DIR_EXAMPLES="$WORK_DIR/examples"
  mkdir -p "$WORK_DIR_EXAMPLES"
  local dir rel dest
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

  # Stage the test-13 single-use rotation realm. Lives outside examples/
  # because it ships only for this hardfork — the bootstrap MsgRun puts it
  # in AllowedDAOs at lock time, a phase-2 MsgCall to Rotate() swaps the
  # sole T1 from manfred to the test-13 T1, then Rotate self-ejects.
  # LoadPackagesFromDir topo-sorts on the gnomod.toml dependency graph, so
  # rotate's addpkg lands after gov/dao + memberstore even though it's
  # copied here separately. See packages/rotate/rotate.gno
  # for the design.
  local ROTATE_PKG_SRC="$SCRIPT_DIR/packages/rotate"
  local ROTATE_PKG_DEST="$WORK_DIR_EXAMPLES/gno.land/r/test13/rotate"
  print_substep "1.4.4" "Staging single-use rotation realm at ${ROTATE_PKG_DEST#"$WORK_DIR_EXAMPLES"/}"
  mkdir -p "$ROTATE_PKG_DEST"
  find "$ROTATE_PKG_SRC" -maxdepth 1 -type f -exec cp {} "$ROTATE_PKG_DEST/" \;

  print_substep "1.4.5" "Creating deployer key..."
  printf '%s\n\n' "$DEPLOYER_MNEMONIC" | run "$GNOKEY_BIN" add --recover GenesisDeployer --home "$WORK_DIR_GNOKEY_HOME" --insecure-password-stdin 2>&1 | sed 's/^/    /'

  print_substep "1.4.6" "Generating empty genesis..."
  run "$GNOGENESIS_BIN" generate -chain-id "$CHAIN_ID" -genesis-time "$GENESIS_TIME" --output-path "$WORK_DIR_GENESIS" 2>&1 | sed 's/^/    /'

  print_substep "1.4.7" "Adding $pkg_count packages to genesis..."
  echo "" | run "$GNOGENESIS_BIN" txs add packages "$WORK_DIR_EXAMPLES" -gno-home "$WORK_DIR_GNOKEY_HOME" -key-name GenesisDeployer --genesis-path "$WORK_DIR_GENESIS" --insecure-password-stdin 2>&1 | sed 's/^/    /'

  print_substep "1.4.8" "Exporting txs..."
  run "$GNOGENESIS_BIN" txs export "$WORK_DIR_GENESIS_TXS" --genesis-path "$WORK_DIR_GENESIS" 2>&1 | sed 's/^/    /'

  # ---- Step 5: Generate bootstrap MsgRun tx (transactions/base/bootstrap/)

  print_phase_header 1 5 "$TOTAL_STEPS" "Generate bootstrap MsgRun tx (transactions/base/bootstrap/)"

  local BOOTSTRAP_DIR="$SCRIPT_DIR/transactions/base/bootstrap"
  local BOOTSTRAP_JSONL="$WORK_DIR/bootstrap_tx.jsonl"

  print_substep "1.5.1" "Building AnnotatedTx from $BOOTSTRAP_DIR/..."
  : >"$BOOTSTRAP_JSONL"
  txn_dir_to_jsonl "$BOOTSTRAP_DIR" "$BOOTSTRAP_JSONL"

  # `txs add sheets` consumes plain TxWithMetadata (no reason field); strip it.
  local BOOTSTRAP_TX_FILE="$WORK_DIR/bootstrap_tx_stripped.jsonl"
  jq -c 'del(.reason)' "$BOOTSTRAP_JSONL" >"$BOOTSTRAP_TX_FILE"

  print_substep "1.5.2" "Adding bootstrap tx to genesis..."
  run "$GNOGENESIS_BIN" txs add sheets "$BOOTSTRAP_TX_FILE" --genesis-path "$WORK_DIR_GENESIS" 2>&1 | sed 's/^/    /'
  cat "$BOOTSTRAP_TX_FILE" >>"$WORK_DIR_GENESIS_TXS"
  verify_checksum "$WORK_DIR_GENESIS_TXS"

  local tx_count
  tx_count=$(wc -l <"$WORK_DIR_GENESIS_TXS" | tr -d ' ')
  print_substep "1.5.3" "Total txs so far: $tx_count"

  # ---- Step 6: Calculate deployer balances
  # Same approach as gnoland1's gen-genesis.sh: spin up a temp node, pre-fund
  # every creator/caller address with $INITIAL_BALANCE, let the genesis txs
  # burn through fees, then query remaining balances. The amount actually
  # spent is what we need to credit each deployer in the real genesis so
  # their balance lands at zero post-genesis (matching gnoland1's "deployer
  # costs are exact, no leftover funds" invariant).
  #
  # Run twice for safety:
  #   run 1: measure actual consumption with over-provisioned balances
  #   run 2: verify the measured balances land everyone at zero
  # If run 2 disagrees, something is non-deterministic and we abort.

  print_phase_header 1 6 "$TOTAL_STEPS" "Calculate deployer balances"

  local BALANCES_TMP_DIR="$WORK_DIR/balances-work"
  local BALANCES_TMP_FILE="$BALANCES_TMP_DIR/balances.txt"
  local BALANCES_TMP_GNOLAND_DATA="$BALANCES_TMP_DIR/gnoland-data"
  local BALANCES_TMP_GNOLAND_LOG="$BALANCES_TMP_DIR/node.log"
  local BALANCES_TMP_GENESIS="$BALANCES_TMP_DIR/genesis.json"
  local BALANCES_TMP_CREATOR_ADDRESSES="$BALANCES_TMP_DIR/gen-creators.txt"
  local INITIAL_BALANCE=1000000000000000
  local NODE_TIMEOUT=120

  pick_free_port() {
    python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()'
  }

  local NODE_RPC_PORT NODE_P2P_PORT NODE_RPC_ADDR
  NODE_RPC_PORT=$(pick_free_port)
  NODE_P2P_PORT=$((NODE_RPC_PORT + 1))
  NODE_RPC_ADDR="127.0.0.1:$NODE_RPC_PORT"

  rm -rf "$BALANCES_TMP_DIR"
  mkdir -p "$BALANCES_TMP_DIR"

  print_substep "1.6.1" "Extracting creator addresses..."
  grep -oE '"(creator|caller)":"[^"]*"' "$WORK_DIR_GENESIS_TXS" |
    sed 's/"creator":"//;s/"caller":"//;s/"//g' |
    sort -u >"$BALANCES_TMP_CREATOR_ADDRESSES"
  local addr_count
  addr_count=$(wc -l <"$BALANCES_TMP_CREATOR_ADDRESSES" | tr -d ' ')
  print_substep "1.6.2" "Found $addr_count unique creator/caller addresses"

  print_substep "1.6.3" "Generating over-provisioned balances..."
  local addr
  while IFS= read -r addr; do
    echo "${addr}=${INITIAL_BALANCE}ugnot" >>"$BALANCES_TMP_FILE"
  done <"$BALANCES_TMP_CREATOR_ADDRESSES"

  # Helper: spin up a temp node with the current genesis + balance sheet.
  # Sets NODE_PID and NODE_RPC_ADDR; aborts if the node doesn't come up in
  # NODE_TIMEOUT seconds.
  start_temp_node() {
    local run_label="$1"

    rm -rf "$BALANCES_TMP_GNOLAND_DATA" "$BALANCES_TMP_GENESIS"
    NODE_RPC_PORT=$(pick_free_port)
    NODE_P2P_PORT=$((NODE_RPC_PORT + 1))
    NODE_RPC_ADDR="127.0.0.1:$NODE_RPC_PORT"

    run "$GNOGENESIS_BIN" generate -chain-id "$CHAIN_ID" -genesis-time "$(date +%s)" -output-path "$BALANCES_TMP_GENESIS"
    run "$GNOGENESIS_BIN" txs add sheets "$WORK_DIR_GENESIS_TXS" -genesis-path "$BALANCES_TMP_GENESIS"
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
      if command -v curl >/dev/null 2>&1; then
        if curl -sf "http://$NODE_RPC_ADDR/status" >/dev/null 2>&1; then
          printf "  Node ready (%ss)\n" "$elapsed"
          return
        fi
      else
        # wget probe — discard body, exit 0 only on HTTP 200.
        if wget -q -O /dev/null "http://$NODE_RPC_ADDR/status" 2>/dev/null; then
          printf "  Node ready (%ss)\n" "$elapsed"
          return
        fi
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
  print_substep "1.6.4" "Querying remaining balances..."
  rm -f "$BALANCES_TMP_FILE"
  local remaining final
  while IFS= read -r addr; do
    remaining=$(query_balance "$addr")
    final=$((INITIAL_BALANCE - remaining))
    printf "    %s = %s ugnot\n" "$addr" "$final"
    echo "${addr}=${final}ugnot" >>"$BALANCES_TMP_FILE"
  done <"$BALANCES_TMP_CREATOR_ADDRESSES"
  stop_temp_node

  start_temp_node "run 2: verify zero balances"
  print_substep "1.6.5" "Verifying all balances are zero..."
  local all_zero=true
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
  print_substep "1.6.6" "All balances zero — deployer costs verified"
  cp "$BALANCES_TMP_FILE" "$WORK_DIR_DEPLOYER_BALANCES"

  # ---- Step 7: Add the initial validator set to the genesis file
  # Done before balances — gnogenesis is O(filesize) per call, so adding
  # validators while the file is still small saves ~6 minutes.

  print_phase_header 1 7 "$TOTAL_STEPS" "Add the initial validator set"

  local validator name power address pub_key
  for validator in "${INITIAL_VALSET[@]}"; do
    read -r name power address pub_key <<<"$validator"
    printf "  %s (power=%s, %s)\n" "$name" "$power" "$address"
    run "$GNOGENESIS_BIN" validator add -name "$name" -power "$power" -address "$address" -pub-key "$pub_key" --genesis-path "$WORK_DIR_GENESIS"
  done

  # ---- Step 8: Generate valoper-seed migration .jsonl + add canonical balances
  # Builds a CSV from (INITIAL_VALSET, INITIAL_VALSET_OPERATORS) and runs
  # `gnogenesis fork valoper-seed` to produce a deterministic .jsonl of
  # genesis-mode valopers.Register MsgCalls. Phase 2 passes this file via
  # --migration-tx so each founding validator gets a valoper profile + v3
  # cache entry at genesis-mode replay. Without it the chain boots with
  # orphan validators (no operator-keyed management plane) and gnoland's
  # InitChainer-side AssertGenesisValopersConsistent fires (#5701).

  print_phase_header 1 8 "$TOTAL_STEPS" "Generate valoper-seed migration .jsonl"

  if [ "${#INITIAL_VALSET_OPERATORS[@]}" -ne "${#INITIAL_VALSET[@]}" ]; then
    die "INITIAL_VALSET_OPERATORS length (${#INITIAL_VALSET_OPERATORS[@]}) must match INITIAL_VALSET length (${#INITIAL_VALSET[@]})"
  fi

  print_substep "1.8.1" "Building CSV from INITIAL_VALSET + INITIAL_VALSET_OPERATORS..."
  local i _power _address op_addr
  {
    echo "operator_addr,signing_pubkey,moniker,description,server_type"
    for i in "${!INITIAL_VALSET[@]}"; do
      read -r name _power _address pub_key <<<"${INITIAL_VALSET[$i]}"
      op_addr="${INITIAL_VALSET_OPERATORS[$i]}"
      # description and server_type are templates — edit if a specific
      # founder needs different metadata. Description must be non-empty
      # and <=2048 chars; server_type ∈ {cloud, on-prem, data-center}.
      printf '%s,%s,%s,test-13 founding validator (%s),cloud\n' \
        "$op_addr" "$pub_key" "$name" "$name"
    done
  } >"$WORK_DIR_VALOPER_CSV"

  print_substep "1.8.2" "Running gnogenesis fork valoper-seed..."
  # --caller is the fee payer for each Register MsgCall (1 ugnot fee — see the
  # Coin amino zero-collapse rationale in valoper_seed.go). We use the GovDAO
  # T1 multisig because it has 119M GNOT from the gnoland1 airdrop, so the 2
  # ugnot total fee is covered without pre-funding any other address. The
  # operator from the CSV row is still passed in MsgCall.Args[3], so each
  # operator gets registered correctly; the squat guard check
  # (caller==operator) is bypassed at genesis-mode (ChainHeight()==0).
  run "$GNOGENESIS_BIN" fork valoper-seed \
    --csv "$WORK_DIR_VALOPER_CSV" \
    --output "$WORK_DIR_VALOPER_SEED" \
    --caller g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh 2>&1 | sed 's/^/    /'

  verify_checksum "$WORK_DIR_VALOPER_SEED"
  print_substep "1.8.3" "-> $WORK_DIR_VALOPER_SEED"

  # Add gnoland1-canonical balances (deployer + airdrop). Faucets are
  # deliberately appended in step 9 instead — see step 9 for the
  # SignerInfo-ordering rationale.

  print_substep "1.8.4" "Adding deployer balances to genesis..."
  run "$GNOGENESIS_BIN" balances add -balance-sheet "$WORK_DIR_DEPLOYER_BALANCES" --genesis-path "$WORK_DIR_GENESIS" >/dev/null

  if [ -f "$AIRDROP_BALANCES_GZ" ]; then
    print_substep "1.8.5" "Using cached airdrop balances: $AIRDROP_BALANCES_GZ"
  else
    print_substep "1.8.5" "Downloading airdrop balances from $BALANCES_GZ_URL..."
    download_url "$BALANCES_GZ_URL" "$AIRDROP_BALANCES_GZ"
  fi
  verify_checksum "$AIRDROP_BALANCES_GZ"
  gunzip_to_stdout "$AIRDROP_BALANCES_GZ" >"$AIRDROP_BALANCES_TXT"

  # Merge airdrop with deployer balances by summing collisions. If an address
  # appears in both the deployer balance sheet (fees consumed in genesis-mode
  # txs) and the airdrop snapshot, `gnogenesis balances add` would otherwise
  # replace the deployer entry with the airdrop entry — losing the deployer's
  # residual / overwriting it. Summing preserves both contributions.
  local AIRDROP_MERGED_TXT="$WORK_DIR/airdrop_merged.txt"
  awk -F= '
    function strip_ugnot(s,    v) {
      v=s
      if (v !~ /ugnot$/) { print "error: non-ugnot balance: " $0 > "/dev/stderr"; exit 1 }
      sub(/ugnot$/, "", v)
      return v+0
    }
    FNR==NR { deployer[$1]=strip_ugnot($2); next }
    { addr=$1; amt=strip_ugnot($2)
      if (addr in deployer) amt+=deployer[addr]
      print addr "=" amt "ugnot" }
  ' "$WORK_DIR_DEPLOYER_BALANCES" "$AIRDROP_BALANCES_TXT" >"$AIRDROP_MERGED_TXT"
  local collision_count
  collision_count=$(awk -F= 'FNR==NR{d[$1]=1;next} $1 in d{c++} END{print c+0}' \
    "$WORK_DIR_DEPLOYER_BALANCES" "$AIRDROP_BALANCES_TXT")
  if [ "$collision_count" -gt 0 ]; then
    print_substep "1.8.6" "Merged $collision_count deployer/airdrop collision(s)"
  fi

  local airdrop_count
  airdrop_count=$(wc -l <"$AIRDROP_MERGED_TXT" | tr -d ' ')
  print_substep "1.8.7" "Adding $airdrop_count airdrop balances to genesis..."
  run "$GNOGENESIS_BIN" balances add -balance-sheet "$AIRDROP_MERGED_TXT" --genesis-path "$WORK_DIR_GENESIS" >/dev/null

  # ---- Step 9: Append faucet balances + verify
  # `gnogenesis balances add` sorts state.Balances by Address.Compare. If we
  # had added faucets in step 8, they would land at sort positions
  # interspersed with the airdrop, shifting every airdrop entry that sorts
  # after them by +1 per shift. That breaks the historical-tx SignerInfo
  # invariant: validateSignerInfo (gno.land/pkg/gnoland/app.go:766)
  # requires that state.Balances[i].Address matches whatever the cached
  # txs.jsonl SignerInfo claims for accNum i — and those numbers come
  # from gnoland1's launch ordering (e.g. manfred at 3,096,261).
  #
  # Note: the 7 NON-manfred deployer addresses ARE in step 8 — they were
  # also in gnoland1's state.Balances at their natural sort positions
  # (gnoland1 used the same deployer mnemonic, so the addresses are
  # byte-identical). That keeps test-13's state.Balances[0..3,262,513]
  # byte-identical to gnoland1's, so all SignerInfo account numbers
  # ≤ 3,262,513 line up.
  #
  # The 10 faucets land at state.Balances[3,262,514..3,262,523]. Empirical
  # scan of the cached txs.jsonl confirms no SignerInfo entry claims an
  # account number in that range, so this is collision-free.
  #
  # Mechanics: build a throwaway genesis containing only the 10 faucets.
  # `gnogenesis balances add` serializes it with the same amino indenter
  # the real genesis uses, so the 10 balance lines come out in the exact
  # format the real genesis already uses. We extract those lines and
  # splice them into the real genesis's state.Balances array right before
  # its closing bracket. No JSON parser involved on the 186 MB file — only
  # amino's own output is consumed, so the resulting SHA is byte-identical
  # to running this whole step through a programmatic amino round-trip.
  #
  # Side note: this is the LAST step that touches state.Balances. Any
  # subsequent `gnogenesis balances add` would re-sort and undo the
  # splice, so the verify call at the end must not modify balances.

  print_phase_header 1 9 "$TOTAL_STEPS" "Append faucet balances + verify genesis"

  # 9.1 Build the faucet balance sheet inline.
  local FAUCET_BALANCES_FILE="$WORK_DIR/faucet_balances.txt"
  : >"$FAUCET_BALANCES_FILE"
  for addr in "${FAUCET_ADDRESSES[@]}"; do
    echo "${addr}=${FAUCET_BALANCE}ugnot" >>"$FAUCET_BALANCES_FILE"
  done
  print_substep "1.9.1" "Faucet balance sheet built (${#FAUCET_ADDRESSES[@]} entries, $FAUCET_BALANCE ugnot each)"

  # 9.2 Throwaway genesis with ONLY the 10 faucets. The amino indenter is
  # the same one the real genesis uses, so the resulting balance lines are
  # format-compatible.
  local FAUCET_TMP_GENESIS="$WORK_DIR/faucet_tmp_genesis.json"
  run "$GNOGENESIS_BIN" generate -chain-id "$CHAIN_ID" -genesis-time "$GENESIS_TIME" --output-path "$FAUCET_TMP_GENESIS" 2>&1 | sed 's/^/    /'
  run "$GNOGENESIS_BIN" balances add -balance-sheet "$FAUCET_BALANCES_FILE" --genesis-path "$FAUCET_TMP_GENESIS" >/dev/null

  # 9.3 Extract the 10 balance lines from the throwaway genesis. The
  # state.Balances array is delimited by `    "balances": [` and `    ],`
  # (4-space indent — `balances` is one level below the genesis root). The
  # 10 lines come out preserving amino's trailing-comma rule: 9 entries
  # with a comma, the last without.
  local FAUCET_EXTRAS_FILE="$WORK_DIR/faucet_extras.lines"
  sed -n '/^    "balances": \[$/,/^    \],*$/{
      /^    "balances": \[$/d
      /^    \],*$/d
      p
  }' "$FAUCET_TMP_GENESIS" >"$FAUCET_EXTRAS_FILE"
  local extras_count
  extras_count=$(wc -l <"$FAUCET_EXTRAS_FILE" | tr -d ' ')
  if [ "$extras_count" -ne "${#FAUCET_ADDRESSES[@]}" ]; then
    die "extracted $extras_count balance lines from throwaway genesis, expected ${#FAUCET_ADDRESSES[@]}"
  fi
  print_substep "1.9.2" "Extracted $extras_count amino-formatted balance lines from throwaway genesis"

  # 9.4 Splice the 10 lines into the real genesis. Awk passes the file
  # through line-by-line and, while inside state.Balances, buffers the
  # previously-seen entry. When it reaches the closing bracket it:
  #   - emits the buffered entry with a comma appended (it was the last
  #     entry before the splice, so previously had no comma);
  #   - emits the 10 extras verbatim (the last one has no comma, which
  #     matches amino's output for the new last-of-array);
  #   - emits the closing bracket itself.
  #
  # Any other "balances" key in the file (none today — verified via
  # `grep -c '"balances"' genesis.json`) would not match: the regex is
  # anchored to exactly 4 spaces of indent and the `[$` / `],?$` shape.
  print_substep "1.9.3" "Splicing into state.Balances tail..."
  awk -v EXTRAS="$FAUCET_EXTRAS_FILE" '
BEGIN { in_bal = 0; prev = "" }
{
  if (in_bal && $0 ~ /^    \],?$/) {
    if (prev != "") print prev ","
    while ((getline ext < EXTRAS) > 0) print ext
    close(EXTRAS)
    in_bal = 0
    prev = ""
    print
    next
  }
  if (in_bal) {
    if (prev != "") print prev
    prev = $0
    next
  }
  if ($0 ~ /^    "balances": \[$/) { in_bal = 1 }
  print
}
' "$WORK_DIR_GENESIS" >"$WORK_DIR_GENESIS.spliced"
  mv "$WORK_DIR_GENESIS.spliced" "$WORK_DIR_GENESIS"
  print_substep "1.9.4" "Splice complete"

  # 9.5 Verify the generated genesis file.
  print_substep "1.9.5" "Running gnogenesis verify..."
  run "$GNOGENESIS_BIN" verify -genesis-path "$WORK_DIR_GENESIS"

  # Final move into place + checksum locks.
  mv "$WORK_DIR_GENESIS" "$GENESIS_FILE"
  verify_checksum "$GENESIS_FILE"

  # ---- Phase summary

  local PHASE_END_TS PHASE_DURATION BASE_GENESIS_BYTES VALOPER_SEED_BYTES BASE_GENESIS_SHA
  PHASE_END_TS=$(date +%s)
  PHASE_DURATION=$((PHASE_END_TS - PHASE_START_TS))
  BASE_GENESIS_BYTES=$(file_size "$GENESIS_FILE")
  VALOPER_SEED_BYTES=$(file_size "$WORK_DIR_VALOPER_SEED")
  BASE_GENESIS_SHA=$(sha256_of "$GENESIS_FILE")

  printf "\n--- Phase 1 complete (%s) ---\n" "$(format_duration "$PHASE_DURATION")"
  printf "   %s (%s, sha256=%s)\n" \
    "${GENESIS_FILE#"$TEST13_DIR"/}" \
    "$(format_size "$BASE_GENESIS_BYTES")" \
    "$BASE_GENESIS_SHA"
  printf "   %s (%s)\n" \
    "${WORK_DIR_VALOPER_SEED#"$TEST13_DIR"/}" \
    "$(format_size "$VALOPER_SEED_BYTES")"
}

# =============================================================================
# Phase 2 — apply gnoland1's historical txs + patches + migration txs.
# =============================================================================

run_phase_2() {
  local PHASE_START_TS
  PHASE_START_TS=$(date +%s)
  local TOTAL_STEPS=5

  printf '\n### Phase 2 — Apply historical replay + patches + migrations ###\n'

  # ---- Step 1: Resolve script paths and tooling

  print_phase_header 2 1 "$TOTAL_STEPS" "Resolve script paths and tooling"

  local BASE_GENESIS="$PHASE1_DIR/base-genesis.json"
  local VALOPER_SEED="$PHASE1_DIR/valoper-seed.jsonl"

  local OUT_GENESIS="$PHASE2_DIR/genesis.json"
  local OUT_MIGRATIONS="$PHASE2_DIR/t1-rotation.jsonl"
  local OUT_FORK_TEST_LOG="$PHASE2_DIR/fork-test.log"
  local OUT_TXS_CACHE="$PHASE2_DIR/txs.jsonl"

  print_substep "2.1.1" "TEST13_DIR=$TEST13_DIR"
  print_substep "2.1.2" "REPO_ROOT=$REPO_ROOT"
  print_substep "2.1.3" "PHASE1_DIR=$PHASE1_DIR (read-only — phase-1 outputs)"
  print_substep "2.1.4" "PHASE2_DIR=$PHASE2_DIR"

  mkdir -p "$PHASE2_DIR" "$WORK_DIR_BIN"

  # ---- Step 2: Verify required tools

  print_phase_header 2 2 "$TOTAL_STEPS" "Verify required tools"

  require_tools \
    "shasum|sha256sum" \
    go jq \
    awk sed grep sort tr mv cp ls find wc head tail cut ps

  print_substep "2.2.1" "All required tools present"

  # Pre-flight checks: resolve the effective txs source and verify required inputs exist.

  if [ ! -f "$BASE_GENESIS" ]; then
    die "$BASE_GENESIS not found — run ./gen-genesis.sh --phase 1 first."
  fi

  if [ ! -f "$VALOPER_SEED" ]; then
    printf 'ERROR: %s not found — run ./gen-genesis.sh --phase 1 first.\n' "$VALOPER_SEED" >&2
    printf '       (phase-1 step 8 emits this file: one valopers.Register tx per INITIAL_VALSET entry).\n' >&2
    exit 1
  fi

  # Phase 1 leaves the deployer key in $WORK_DIR_GNOKEY_HOME; if the user
  # ran with a clean work/, that home is gone — abort with a clear
  # message instead of mysteriously failing inside gnokey.
  if [ ! -d "$WORK_DIR_GNOKEY_HOME/data/keys.db" ]; then
    printf 'ERROR: deployer keybase not found at %s/data/keys.db\n' "$WORK_DIR_GNOKEY_HOME" >&2
    printf '       Re-run ./gen-genesis.sh --phase 1 to repopulate it.\n' >&2
    exit 1
  fi

  # Txs source: default to multi-RPC fetch against DEFAULT_TXS_RPC_ENDPOINTS.
  # The user can override by passing --source-txs-jsonl-file or
  # --source-txs-data-dir; in those modes the path/dir must exist.
  if [ "$TXS_SOURCE_COUNT" -eq 0 ]; then
    SOURCE_TXS_RPC="$DEFAULT_TXS_RPC_ENDPOINTS"
  fi
  if [ -n "$SOURCE_TXS_JSONL_FILE" ] && [ ! -f "$SOURCE_TXS_JSONL_FILE" ]; then
    die "--source-txs-jsonl-file points at $SOURCE_TXS_JSONL_FILE which does not exist."
  fi
  if [ -n "$SOURCE_TXS_DATA_DIR" ] && [ ! -d "$SOURCE_TXS_DATA_DIR" ]; then
    die "--source-txs-data-dir points at $SOURCE_TXS_DATA_DIR which does not exist."
  fi

  # ---- Step 3: Build binaries + verify txs source

  print_phase_header 2 3 "$TOTAL_STEPS" "Build binaries + verify txs source"

  if [ "$NO_INSTALL" = true ]; then
    print_substep "2.3.1" "--no-install — reusing $WORK_DIR_BIN"
    local bin
    for bin in "$GNOKEY_BIN" "$GNOGENESIS_BIN"; do
      if [ ! -x "$bin" ]; then
        die "--no-install but $bin not found. Run without --no-install first."
      fi
    done
  else
    print_substep "2.3.1" "Building gnokey..."
    run go build -C "$GNOKEY_CMD" -o "$GNOKEY_BIN" .
    print_substep "2.3.2" "Building gnogenesis..."
    run go build -C "$GNOGENESIS_CMD" -o "$GNOGENESIS_BIN" .
  fi

  # Verify the txs source (file modes only). For --source-txs-rpc and
  # --source-txs-data-dir, gnogenesis enforces --halt-height on the
  # fetch side. For --source-txs-jsonl-file, we sanity-check that the
  # cached archive doesn't extend past HALT_HEIGHT — saves an hour of
  # replay if the cache is from a chain that didn't halt yet. The cache
  # is allowed to fall short of HALT_HEIGHT because a chain can have
  # tx-less blocks at the very end.
  if [ -n "$SOURCE_TXS_JSONL_FILE" ]; then
    local TXS_COUNT MAX_HEIGHT
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
    print_substep "2.3.3" "mode:       jsonl-file ($SOURCE_TXS_JSONL_FILE)"
    print_substep "2.3.4" "txs:        $TXS_COUNT"
    print_substep "2.3.5" "max height: $MAX_HEIGHT (HALT_HEIGHT = $HALT_HEIGHT)"
    if [ "$MAX_HEIGHT" -gt "$HALT_HEIGHT" ]; then
      printf 'ERROR: HALT_HEIGHT=%s but txs.jsonl max BlockHeight=%s (cache extends past halt).\n' \
        "$HALT_HEIGHT" "$MAX_HEIGHT" >&2
      printf '       Update the HALT_HEIGHT constant in this script or replace the cached jsonl.\n' >&2
      exit 1
    fi
    verify_checksum "$SOURCE_TXS_JSONL_FILE"
  elif [ -n "$SOURCE_TXS_RPC" ]; then
    print_substep "2.3.3" "mode:       rpc ($SOURCE_TXS_RPC)"
    print_substep "2.3.4" "(gnogenesis fork generate enforces halt-height during fetch)"
  else
    print_substep "2.3.3" "mode:       data-dir ($SOURCE_TXS_DATA_DIR)"
    print_substep "2.3.4" "(gnogenesis fork generate enforces halt-height during read)"
  fi

  # ---- Step 4: Build migration .jsonl + assemble final genesis
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
  # only for the genesis-mode txs of phase 1 (step 6 calculates the exact
  # fee total and credits it). They land at zero after phase 1, so a
  # migration-tx fee of 1ugnot from the deployer would fail at
  # DeductFees with std.InsufficientFundsError. Admin-as-caller dodges
  # this without needing an extra balance allocation.

  print_phase_header 2 4 "$TOTAL_STEPS" "Build migration .jsonl + assemble final genesis"

  : >"$OUT_MIGRATIONS"
  txn_dir_to_jsonl "$SCRIPT_DIR/transactions/migration/rotate-call" "$OUT_MIGRATIONS"
  print_substep "2.4.1" "MsgCall gno.land/r/test13/rotate.Rotate  caller=$NAMES_ADMIN"

  txn_dir_to_jsonl "$SCRIPT_DIR/transactions/migration/names-enable" "$OUT_MIGRATIONS"
  print_substep "2.4.2" "MsgCall gno.land/r/sys/names.Enable  caller=$NAMES_ADMIN"

  local mig_lines
  mig_lines=$(wc -l <"$OUT_MIGRATIONS" | tr -d ' ')
  print_substep "2.4.3" "-> $OUT_MIGRATIONS ($mig_lines migration txs)"
  verify_checksum "$OUT_MIGRATIONS"

  # Assemble final genesis via gnogenesis fork generate.
  print_substep "2.4.4" "Running gnogenesis fork generate..."

  local GEN_ARGS=(
    fork generate
    --original-chain-id "$ORIGINAL_CHAIN_ID"
    --chain-id "$CHAIN_ID"
    --halt-height "$HALT_HEIGHT"
    --migration-tx "$VALOPER_SEED"
    --migration-tx "$OUT_MIGRATIONS"
    --output "$OUT_GENESIS"
    --txs-output "$OUT_TXS_CACHE"
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

  # Patches: each group dir under transactions/patched/ becomes one --patch-txs
  # jsonl. Per-tx <group>/h<height>/{meta.json + optional body.gno or pkg/} are
  # converted by txn_dir_to_jsonl and concatenated per group. gnogenesis fork
  # generate validates patches cross-file (dupes, unmatched keys) and fails
  # fast on conflicts.
  local PATCHES_DIR="$SCRIPT_DIR/transactions/patched"
  local PATCHES_WORK="$PHASE2_DIR/patches-generated"
  rm -rf "$PATCHES_WORK"
  mkdir -p "$PATCHES_WORK"
  local PATCH_COUNT=0 PATCH_ENTRY_COUNT=0
  shopt -s nullglob
  local group_dir tx_dir group_jsonl
  for group_dir in "$PATCHES_DIR"/*/; do
    group=$(basename "$group_dir")
    group_jsonl="$PATCHES_WORK/$group.jsonl"
    : >"$group_jsonl"
    for tx_dir in "$group_dir"*/; do
      [ -f "$tx_dir/meta.json" ] || continue
      txn_dir_to_jsonl "$tx_dir" "$group_jsonl"
      PATCH_ENTRY_COUNT=$((PATCH_ENTRY_COUNT + 1))
    done
    if [ -s "$group_jsonl" ]; then
      GEN_ARGS+=(--patch-txs "$group_jsonl")
      PATCH_COUNT=$((PATCH_COUNT + 1))
    fi
  done
  shopt -u nullglob
  if [ "$PATCH_COUNT" -gt 0 ]; then
    print_substep "2.4.5" "Loading $PATCH_ENTRY_COUNT patch(es) in $PATCH_COUNT group(s) from $PATCHES_DIR/"
  fi

  run "$GNOGENESIS_BIN" "${GEN_ARGS[@]}"

  verify_checksum "$OUT_GENESIS"
  local OUT_GENESIS_SHA OUT_GENESIS_BYTES
  OUT_GENESIS_SHA=$(sha256_of "$OUT_GENESIS")
  OUT_GENESIS_BYTES=$(file_size "$OUT_GENESIS")
  print_substep "2.4.6" "sha256: $OUT_GENESIS_SHA"
  print_substep "2.4.7" "-> $OUT_GENESIS ($(format_size "$OUT_GENESIS_BYTES"))"
  print_substep "2.4.8" "-> $OUT_MIGRATIONS (kept for audit)"

  # ---- Step 5: Audit + final move
  # Replays the assembled genesis in-process via `gnogenesis fork test`.
  # Any tx failure aborts the build with a top-10-by-frequency print.

  if [ "$SKIP_AUDIT" = true ]; then
    print_phase_header 2 5 "$TOTAL_STEPS" "Audit skipped (--skip-audit) + final move"
  else
    print_phase_header 2 5 "$TOTAL_STEPS" "Audit + final move"

    print_substep "2.5.1" "Running gnogenesis fork test --verbose --skip-failing-genesis-txs..."
    print_substep "2.5.2" "(full output: $OUT_FORK_TEST_LOG)"

    # --skip-failing-genesis-txs makes fork test exit 0 even on tx failures;
    # we parse the verbose log ourselves to decide pass/fail. Suppress the
    # binary's stdout summary — we'll print our own.
    "$GNOGENESIS_BIN" fork test \
      --genesis "$OUT_GENESIS" \
      --verbose \
      --skip-failing-genesis-txs \
      --timeout 1h \
      >"$OUT_FORK_TEST_LOG" 2>&1 &
    local FORK_TEST_PID=$!

    # Spinner: fork test takes minutes; show progress every 5s.
    local elapsed
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
    # We extract one "<Type>: <key>" line per failure into FAIL_LINES_FILE
    # for human-readable frequency counting on abort.
    local FAIL_LINES_FILE="$PHASE2_DIR/fork-test-failures.txt"
    grep -E "^Data: " "$OUT_FORK_TEST_LOG" | sed -E '
      s/^Data: (vm\.TypeCheckError)\{[^"]*Errors:\[\]string\{"([^"]+)".*/\1: \2/
      s/^Data: (std\.InsufficientFeeError)\{.*/\1: insufficient fee/
      s/^Data: ([A-Za-z][A-Za-z0-9_.]+)\{.*/\1: (no detail)/
    ' >"$FAIL_LINES_FILE" 2>/dev/null || true
    local TOTAL_FAILS
    TOTAL_FAILS=$(wc -l <"$FAIL_LINES_FILE" | tr -d ' ')

    if [ "$TOTAL_FAILS" -eq 0 ]; then
      print_substep "2.5.3" "No failed txs."
    else
      printf "\n  %s failed txs — top-10 by frequency:\n" "$TOTAL_FAILS"
      sort "$FAIL_LINES_FILE" | uniq -c | sort -rn | head -10 |
        sed 's/^/    /'
      printf "\n  Full per-failure list: %s\n" "$FAIL_LINES_FILE"
      printf "  Full fork-test log:   %s\n" "$OUT_FORK_TEST_LOG"
      exit 1
    fi
  fi

  # Final move into place. Verify the source artifact then mv to root.
  # Re-verify after mv to lock the root-level genesis.json — same content,
  # new path, so the CHECKSUMS entry is keyed on the new relative path.
  print_substep "2.5.4" "Moving $OUT_GENESIS -> $FINAL_GENESIS"
  mv "$OUT_GENESIS" "$FINAL_GENESIS"
  verify_checksum "$FINAL_GENESIS"
  local FINAL_GENESIS_SHA FINAL_GENESIS_BYTES
  FINAL_GENESIS_SHA=$(sha256_of "$FINAL_GENESIS")
  FINAL_GENESIS_BYTES=$(file_size "$FINAL_GENESIS")

  # Provenance report: counts per Source value + per-tx reasons for the
  # patched + migration entries. Lets the operator + validators eyeball the
  # final genesis composition without grep'ing the 192MB blob.
  print_substep "2.5.5" "Provenance report (gnogenesis fork inspect)..."
  run "$GNOGENESIS_BIN" fork inspect "$FINAL_GENESIS" 2>&1 | sed 's/^/    /'

  # ---- Phase summary

  local PHASE_END_TS PHASE_DURATION
  PHASE_END_TS=$(date +%s)
  PHASE_DURATION=$((PHASE_END_TS - PHASE_START_TS))

  printf "\n--- Phase 2 complete (%s) ---\n" "$(format_duration "$PHASE_DURATION")"
  printf "   %s (%s, sha256=%s)\n" \
    "${FINAL_GENESIS#"$TEST13_DIR"/}" \
    "$(format_size "$FINAL_GENESIS_BYTES")" \
    "$FINAL_GENESIS_SHA"
  printf "   %s (kept for audit)\n" "${OUT_MIGRATIONS#"$TEST13_DIR"/}"
  if [ "$SKIP_AUDIT" != true ]; then
    printf "   %s (kept for audit)\n" "${OUT_FORK_TEST_LOG#"$TEST13_DIR"/}"
  fi
}

# =============================================================================
# Main dispatch.
# =============================================================================

PIPELINE_START_TS=$(date +%s)

printf '\n### test-13 hardfork genesis build (phase=%s) ###\n' "$PHASE"

if [ "$PHASE" = "1" ] || [ "$PHASE" = "all" ]; then
  run_phase_1
fi

if [ "$PHASE" = "2" ] || [ "$PHASE" = "all" ]; then
  if [ "$PHASE" = "all" ]; then
    printf '\n### Phase 1 -> Phase 2 handoff ###\n'
  fi
  run_phase_2
fi

PIPELINE_END_TS=$(date +%s)
PIPELINE_DURATION=$((PIPELINE_END_TS - PIPELINE_START_TS))

if [ -f "$FINAL_GENESIS" ] && [ "$PHASE" != "1" ]; then
  FINAL_SHA=$(sha256_of "$FINAL_GENESIS")
  FINAL_BYTES=$(file_size "$FINAL_GENESIS")
  printf '\n### test-13 build complete: genesis.json (%s, sha256=%s) ###\n' \
    "$(format_size "$FINAL_BYTES")" "$FINAL_SHA"
fi
printf '    total pipeline time: %s\n' "$(format_duration "$PIPELINE_DURATION")"
