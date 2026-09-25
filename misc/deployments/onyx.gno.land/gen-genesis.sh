#!/usr/bin/env bash
# gen-genesis.sh — onyx genesis builder (single-file pipeline).
#
# onyx is the testnet on the mainnet line: it runs mainnet's exact binaries
# (the version in UPGRADES.md) and is upgraded whenever mainnet is. Its genesis
# is mainnet's shape with a testnet's money — a fresh chain, no hardfork, no
# historical replay, built from the repo's examples/ tree plus a handful of
# bootstrap txs, in minutes. Derived from misc/deployments/mainnet.gno.land/
# with the independence-day allocation, the §126 transfer lock and the §132
# vesting machinery removed, and four typed balances in their place.
#
# What the genesis contains:
#
#   1. The FILTERED_PACKAGES example set (resolved with transitive deps),
#      addpkg'd by the deterministic GenesisDeployer key — the same set as
#      mainnet's.
#   2. A bootstrap MsgRun (transactions/base/bootstrap/) that seeds the sole
#      GovDAO T1 member (aeddi) and locks dao.UpdateImpl's AllowedDAOs to
#      r/gov/dao/impl/v0. A second MsgRun (transactions/base/users-preregister/)
#      registers the same initial namespaces as mainnet in r/sys/users via the
#      genesis-only path.
#   3. A names.Enable MsgCall (transactions/migration/names-enable/) so
#      namespace enforcement is on from genesis. Enable is gated on the admin
#      address hardcoded in r/sys/names/verifier.gno; the tx's caller field is
#      jq-patched to that address post-sign, which the chain trusts under
#      --skip-genesis-sig-verification.
#   4. A valopers.Register MsgCall (emitted by `gnogenesis fork valoper-seed`
#      from INITIAL_VALSET + INITIAL_VALSET_OPERATORS) so the founding
#      validator has an operator-keyed valoper profile and r/sys/validators/v0
#      can manage the set post-genesis.
#   5. The INITIAL_VALSET as GenesisDoc.Validators (InitChainer seeds
#      valset:current from it, so v0/EndBlocker valset changes work).
#   6. Balances: the FUNDED_ACCOUNTS sheet — the two faucet accounts, aeddi,
#      and the gpao approvals oracle — plus exact-burn funding for every
#      genesis-tx fee payer (measured on a temp node; fee payers land at zero,
#      or at exactly their funded amount if they are also in the sheet, once
#      the genesis txs execute).
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
# Build it from the tree of the version onyx launches on (the chain/onyx tag,
# which sits on that version's commit on chain/mainnet): the package set is
# read from examples/, and master's examples/ already differs from what the
# network runs. Cross-platform: bash 3.2 minimum (macOS default), no GNU-only
# features.

# -u is deliberately absent: on bash 3.2 (macOS default, the floor this script
# targets) expanding an EMPTY array with "${arr[@]}" errors under -u, and both
# RESTRICTED_DENOMS and txn_dir_to_jsonl's args_array are legitimately empty.
set -eo pipefail

# =============================================================================
# Launch parameters — review before each genesis generation.
# =============================================================================

# The chain id the validator key ceremony ran with, and the network id gpao
# signs every approval with: both are sign-bytes, so this is not a free name.
CHAIN_ID=onyx-1 # decided 2026-09-25

# DECIDED (A, 2026-09-26): launch at 2026-09-28T00:00:00Z. Block 1 carries this
# timestamp forever; step 9 asserts the shipped genesis_time is exactly it.
GENESIS_TIME=1790553600 # 2026-09-28T00:00:00Z — decided 2026-09-26

# Packages to include in genesis (resolved with transitive dependencies).
# Use "..." suffix to match all sub-packages.
#
# Identical to mainnet's FILTERED_PACKAGES, on purpose: onyx exists to
# rehearse mainnet's upgrades, so it deploys mainnet's package set. Keep in
# mind when touching it: p/nt/* paths cannot be added post-genesis under
# namespace enforcement, so anything missing here is a relaunch away
# (r/tests/* and r/demo/* arrive via test deps of this set).
FILTERED_PACKAGES=(
  ./gno.land/r/sys/...
  ./gno.land/r/gov/...
  ./gno.land/r/gnoland/blog/...
  ./gno.land/r/gnoland/wugnot/...
  ./gno.land/r/gnoland/coins/...
  ./gno.land/r/gnoland/boards2/...
  ./gno.land/r/gnops/valopers/...
  ./gno.land/p/onbloc/uint256/v0
  ./gno.land/p/onbloc/int256/v0
  ./gno.land/p/onbloc/json/v0
  ./gno.land/r/sys/validators/v0
  ./gno.land/r/nt/grc20reg/v0
  ./gno.land/p/nt/grc20/v0
  ./gno.land/p/nt/grc721/...
)

# Initial onyx validator set. Format: "name power address pub_key".
#
# One founding validator, run by gno-core. Power 60 as on mainnet — 60
# divides evenly many ways, so later valset changes can hand out proportional
# fractions without fractional remainders; with a single validator it is 100%
# either way. The identity was derived from the key at the ceremony and
# checked a second way on 2026-09-25 (address = bech32("g",
# sha256(pubkey)[:20])). The name follows the house <org>-validator-<n> form.
# How the signing key is protected is the operator's own infra, outside this
# genesis.
INITIAL_VALSET=(
  "gno-core-validator-1 60 g19429enajcectdlhwulwza8x2h6mkygu0hpt6y4 gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqq83v3kcqg709z574hwjn2y7fd0mtt4vuxkvpnjeen5uq5gsarf3rtpksl"
)

# Operator address for each INITIAL_VALSET entry (same index). MUST be
# distinct from the signing address — `gnogenesis fork valoper-seed` rejects
# operator==signing_addr to keep signing-key compromise from collapsing into
# operator-slot compromise (see valoper_seed.go).
#
# The operator key is the management plane for the validator: whoever holds
# it can rotate the signing key, edit the valoper profile, and signal opt-out
# via r/gnops/valopers + r/sys/validators/v0. aeddi's operational key, the
# same address as his GovDAO T1 seat — deliberate, and funded below.
INITIAL_VALSET_OPERATORS=(
  "g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m" # gno-core-validator-1 operator (aeddi)
)

# Genesis balances, in the balance-sheet syntax `<address>=<amount>ugnot`.
# Four accounts and nothing else (decided 2026-09-26); every other genesis
# account is a fee payer funded with the exact burn of its genesis txs and
# lands at zero once they execute.
#
# The faucets and aeddi hold 1e18 ugnot each (1 trillion GNOT), the ceiling
# pearl used: it leaves ~9x headroom under the int64 Coin limit per account,
# and step 2 refuses a sheet whose total gets near that limit — Coin.Add
# panics on int64 overflow, and there is no Constitution here to keep the
# supply meaningful. The gpao oracle holds 1,000,000 GNOT: at contribs/gpao's
# default 1000000ugnot fee that is a million approvals, so it never needs a
# top-up (A, 2026-09-26). Addresses from the 2026-09-25 inputs record, each
# derived from its mnemonic in a throwaway keybase and checked against the
# value recorded at generation.
FUNDED_ACCOUNTS=(
  "g1jf3xq3yur9tp9ts0h9qkl9t96kptvzq8cwshyw=1000000000000000000ugnot" # faucet (captcha) dispensing account
  "g14fqsevre6yjkq5frpkf00zs34rdq82huv6ykxv=1000000000000000000ugnot" # faucet-agent dispensing account
  "g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m=1000000000000000000ugnot" # aeddi: sole GovDAO T1 member, validator operator
  "g1yaee4f2qt8yyzse54wq7r897ndupnvxcyl6adz=1000000000000ugnot"       # gpao approvals oracle (1,000,000 GNOT)
)
# Refused above this: the sum of FUNDED_ACCOUNTS plus the measured fee-payer
# burn must stay well under the int64 Coin limit (9,223,372,036,854,775,807).
FUNDED_TOTAL_CEILING=9000000000000000000

# Denominations subject to a transfer lock. EMPTY: onyx transfers are open —
# a faucet has to send, and there is no Constitution §126 to honour on a
# testnet. This is the one place that differs from mainnet's bank params, and
# step 9 asserts the shipped genesis carries no lock and no exemption list.
RESTRICTED_DENOMS=()

# ---- Inert code-submission policy (decided) ----
#
# onyx runs gpao exactly as mainnet does: a post-genesis MsgAddPackage is
# STORED, not executed, and becomes live only when an address in PKG_APPROVERS
# sends MsgEnablePackage for the exact bytes it reviewed. Genesis replay is
# exempt (auth.IsGenesisReplay), so the genesis packages still execute at
# block 1. The guards in step 2 are mainnet's, minus the ones that only exist
# because mainnet has no faucet.
CODE_SUBMISSION_POLICY=inert

# Addresses permitted to send MsgEnablePackage.
#
# Load-bearing: with no approver, every submission parks forever and the chain
# accepts deploys it can never activate (vm.enableBlockedReason ->
# ReasonNoApprovers). Nothing on chain refuses that state, so step 2 refuses
# to build it. Keep this to the gpao oracle key and nothing else: an approver
# can activate any parked package, and this key lives unattended on an
# internet-facing daemon. It is funded in FUNDED_ACCOUNTS above; step 2
# asserts that, since an approver pays gas for every MsgEnablePackage.
PKG_APPROVERS=(
  g1yaee4f2qt8yyzse54wq7r897ndupnvxcyl6adz # gpao approval oracle (onyx key, generated 2026-09-26)
)

# Addresses permitted to send MsgRun. EMPTY MEANS OPEN — the param's zero value
# is "off", so leaving this empty lets anyone execute arbitrary source on a
# chain whose whole submission policy exists to stop exactly that. "inert"
# gates package PERSISTENCE; this is the gate on EXECUTION, and only both
# together give the property the policy is named for.
#
# GovDAO proposal creation is MsgRun-only (a ProposalRequest carries a func
# value, which MsgCall cannot marshal), so every T1 member must appear here or
# governance is unreachable AND the list can never be amended.
#
# DERIVED, not typed: the members are read out of the bootstrap tx and unioned
# in at step 2, so the list cannot drift from what actually gets seeded. On
# onyx that means only aeddi can `maketx run` — the same gate as mainnet, by
# design; a developer who needs it goes into RUN_SUBMITTERS_EXTRA by name.
RUN_SUBMITTERS_INCLUDE_GOVDAO_T1=true

# Addresses that must `maketx run` WITHOUT being GovDAO members. Empty is the
# expected state; add an operator here only with a reason, since MsgRun
# executes arbitrary source.
RUN_SUBMITTERS_EXTRA=()

# OFF, to match mainnet: the charge is a storage-deposit-like fee on parked
# submissions and a testnet is where behaviour should mirror the network it
# rehearses. Unlike mainnet nothing forbids it here (transfers are open), so
# turning it on is a GovDAO param proposal away if parked blobs pile up.
INERT_SUBMISSION_CHARGE=
INERT_CHARGE_COLLECTOR=

# =============================================================================
# Internal — everything below is glue, you shouldn't need to change it.
# =============================================================================

# Deployer key mnemonic (deterministic — used only to sign genesis-mode txs).
# Same as mainnet/gnoland1/test13/topaz so the deployer address is reproducible.
DEPLOYER_MNEMONIC="anchor hurt name seed oak spread anchor filter lesson shaft wasp home improve text behind toe segment lamp turn marriage female royal twice wealth"
DEPLOYER_KEY=GenesisDeployer
# Address derived from DEPLOYER_MNEMONIC. Used as the fee payer for the
# valoper-seed Register tx; the balance-measurement step funds it exactly.
DEPLOYER_ADDR=g1edq4dugw0sgat4zxcw9xardvuydqf6cgleuc8p

# r/sys/names admin: hardcoded in examples/gno.land/r/sys/names/verifier.gno
# — the gnolang/multisigs [govdao] 4-of-7 multisig, the same value mainnet
# ships. names.Enable's admin check reads runtime.PreviousRealm().Address();
# under --skip-genesis-sig-verification the chain trusts the MsgCall.Caller
# field as the EOA, so jq-patching caller to this address makes Enable's gate
# pass. The private key is not needed at build time. Step 2 reads the admin
# out of the tree and refuses a stale copy here; step 6 checks the tx's
# caller_override matches too. The address also owns r/gnoland/blog and
# r/gnoland/boards2/v0 on this chain; with a faucet, it can be funded whenever
# someone needs to act as that owner, so it holds no genesis balance.
NAMES_ADMIN=g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx

# ---- Locked sha256 hashes.
#
# Format (matches `shasum -a 256` / `sha256sum` output exactly):
#   <sha256>  <path-relative-to-onyx.gno.land>
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
  cat <<'EOF_CHECKSUMS'
# LOCKED 2026-09-26 — the onyx-1 launch build (every launch value final),
# built from the chain/mainnet tree at the v1.5.0 code (2ed33ea9a). Any run
# producing different bytes for these artifacts fails loudly; a run from
# master fails on packages.gen.txt by design, since master's examples/ has
# moved on from what the network runs.
0f58018876aa393456190c5236ad8b66b1c7a0e1d25ab31674b26f5d8d14a960  work/packages.gen.txt
623202193678c47d349fff159ab74dcbc77baae8fd52f23be9966135e4d90d95  work/valoper-seed.jsonl
e466c739ab146fbb4a325181565ce8ce65d86425d2b3477d7cba482a2053ad63  work/genesis_txs.jsonl
40b9d9b0cb92ba5df3f50bba7d4fdbad105ec932d31204cb78a0b07caf3fe7e5  work/deployers_balances.txt
4b006fd7ccdec052865accc84dd29b2b76f8b57b2560789a15eedaa88f0e26c5  genesis.json
EOF_CHECKSUMS
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
    awk | sed | grep | sort | tr | mv | cp | ls | find | wc | head | tail | cut | comm | uniq)
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
# <path> relative to DEPLOY_DIR) in the inline CHECKSUMS_DATA heredoc,
# and one of:
#   - hash matches               → silent OK
#   - hash differs               → FAIL with expected vs got
#   - key not listed             → print computed sha256 + the line to append
CHECKSUMS_LOCKED=0
CHECKSUMS_UNLOCKED=0
CHECKSUMS_UNLOCKED_LINES=""

verify_checksum() {
  local path="$1"
  if [ -z "${DEPLOY_DIR:-}" ]; then
    die "verify_checksum: DEPLOY_DIR not set"
  fi
  if [ ! -f "$path" ]; then
    die "verify_checksum: $path does not exist"
  fi

  local rel="${2:-${path#"$DEPLOY_DIR"/}}"
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
    # Counted so the closing summary can say how much of this build was
    # actually verified: the append lines below scroll past mid-run,
    # interleaved with gnogenesis output, and "no mismatch reported" reads
    # like "locked" when it means "never checked".
    CHECKSUMS_UNLOCKED=$((CHECKSUMS_UNLOCKED + 1))
    CHECKSUMS_UNLOCKED_LINES="$CHECKSUMS_UNLOCKED_LINES$got  $rel"$'\n'
    printf '  [checksum] %s\n' "$rel" >&2
    printf '             not listed in CHECKSUMS_DATA. Append to lock:\n' >&2
    printf '             %s  %s\n' "$got" "$rel" >&2
    return 0
  fi
  CHECKSUMS_LOCKED=$((CHECKSUMS_LOCKED + 1))

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

# sheet_total <path>  →  the ugnot the sheet grants, summed
#
# Summed in bash's 64-bit integers, not awk's doubles: the faucets hold 1e18
# each, far past the 2^53 (9.007e15) where a double stops being exact, and a
# sum of two of them is already a wrong number in awk. Each addition is
# checked for wrap-around — bash wraps silently past 2^63 — so a sheet that
# cannot be summed is refused rather than mis-summed. Pass "-" to sum stdin.
sheet_total() {
  local total=0 amount line
  while IFS= read -r line; do
    [ -z "$line" ] && continue
    amount="${line#*=}"
    amount="${amount%%ugnot*}"
    case "$amount" in
    '' | *[!0-9]*) die "sheet_total: unparseable amount in '$line'" ;;
    esac
    total=$((total + amount))
    if [ "$total" -lt "$amount" ]; then
      die "sheet_total: the sum overflowed a 64-bit integer at '$line' — the genesis supply cannot be represented"
    fi
  done <"$1"
  printf '%d' "$total"
}

# =============================================================================
# Flag parsing.
# =============================================================================

DEBUG=false
NO_INSTALL=false

print_usage() {
  cat <<'EOF_USAGE'
gen-genesis.sh — onyx genesis builder (single-file pipeline).

Usage:
  ./gen-genesis.sh [flags]

Flags:
  --no-install    Reuse previously built binaries in work/bin/.
  --debug         Echo the main pipeline commands before running them.
  -h, --help      Print this help and exit.

Output:
  genesis.json    Final artifact, sha256-locked against the
                  CHECKSUMS_DATA heredoc in this script.

See misc/deployments/onyx.gno.land/README.md for what the genesis
contains.
EOF_USAGE
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
DEPLOY_DIR="$SCRIPT_DIR"
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

printf '\n### onyx genesis build ###\n'

# ---- Code-submission vm params ----
# Applied to BOTH the shipping genesis and the step-8 measurement genesis, from
# ONE function. The params-parity assertion in step 8 requires the two to be
# byte-identical, and two call sites building the same JSON is exactly how they
# drift.
#
# Only non-empty values are written, so an unset list stays at the null the
# generator produced. Writing [] instead would mean the same thing to Params
# but would have to be written identically on both sides to keep parity.
apply_code_submission_params() {
  local genesis="$1" patched="$1.vmparams" approvers_json run_json applied applied_n
  approvers_json=$(printf '%s\n' "${PKG_APPROVERS[@]}" | jq -R -s -c 'split("\n") | map(select(length > 0))')
  run_json=$(printf '%s\n' "${RUN_SUBMITTERS[@]}" | jq -R -s -c 'split("\n") | map(select(length > 0))')
  jq --arg policy "$CODE_SUBMISSION_POLICY" \
    --argjson approvers "$approvers_json" \
    --argjson runners "$run_json" \
    --arg charge "$INERT_SUBMISSION_CHARGE" \
    --arg collector "$INERT_CHARGE_COLLECTOR" \
    '.app_state.vm.params.code_submission_policy = $policy
     | if ($approvers | length) > 0 then .app_state.vm.params.pkg_approvers = $approvers else . end
     | if ($runners | length) > 0 then .app_state.vm.params.run_submitters = $runners else . end
     | if $charge != "" then .app_state.vm.params.inert_submission_charge = $charge else . end
     | if $collector != "" then .app_state.vm.params.inert_charge_collector = $collector else . end' \
    "$genesis" >"$patched"
  mv "$patched" "$genesis"

  # Read back. A jq path typo writes a new key and reports nothing, and the
  # chain would then boot on the generator default -- "permissionless" -- with
  # every guard above having passed.
  applied=$(jq -r '.app_state.vm.params.code_submission_policy' "$genesis")
  if [ "$applied" != "$CODE_SUBMISSION_POLICY" ]; then
    die "code_submission_policy did not apply to $genesis: wanted '$CODE_SUBMISSION_POLICY', genesis has '$applied'"
  fi
  applied_n=$(jq -r '.app_state.vm.params.pkg_approvers | length // 0' "$genesis")
  if [ "$applied_n" -ne "${#PKG_APPROVERS[@]}" ]; then
    die "pkg_approvers did not apply to $genesis: wanted ${#PKG_APPROVERS[@]}, genesis has $applied_n"
  fi
  applied_n=$(jq -r '.app_state.vm.params.run_submitters | length // 0' "$genesis")
  if [ "$applied_n" -ne "${#RUN_SUBMITTERS[@]}" ]; then
    die "run_submitters did not apply to $genesis: wanted ${#RUN_SUBMITTERS[@]}, genesis has $applied_n"
  fi
}

# ---- Step 1: Resolve script paths and tooling

print_step_header 1 "$TOTAL_STEPS" "Resolve script paths and tooling"

GENESIS_FILE="$WORK_DIR/genesis.json"
PACKAGES_GEN_FILE="$WORK_DIR/packages.gen.txt"
GENESIS_TXS_JSONL="$WORK_DIR/genesis_txs.jsonl"
DEPLOYER_BALANCES="$WORK_DIR/deployers_balances.txt"
FUNDED_SHEET="$WORK_DIR/funded_accounts.txt"
VALOPER_CSV="$WORK_DIR/valoper_profiles.csv"
VALOPER_SEED="$WORK_DIR/valoper-seed.jsonl"
# Read at step 2 (T1 funding guard) and added to the genesis at step 5.
BOOTSTRAP_DIR="$SCRIPT_DIR/transactions/base/bootstrap"

print_substep "1.1" "DEPLOY_DIR=$DEPLOY_DIR"
print_substep "1.2" "REPO_ROOT=$REPO_ROOT"
print_substep "1.3" "WORK_DIR=$WORK_DIR"

# ---- Step 2: Verify required tools and launch values

print_step_header 2 "$TOTAL_STEPS" "Verify required tools and launch values"

require_tools \
  "shasum|sha256sum" \
  go jq python3 comm uniq \
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

# ---- Funded accounts ----
# Written to a sheet once, then read by every guard below and by step 8's
# merge, so there is one source for the money in this genesis. Shape-checked
# here, before the expensive build: one `g1<38>=<digits>ugnot` line per
# account, no duplicates, and a total well under the int64 Coin limit.
printf '%s\n' "${FUNDED_ACCOUNTS[@]}" | sed 's/[[:space:]]*#.*$//' | awk 'NF' >"$FUNDED_SHEET"
funded_count=$(wc -l <"$FUNDED_SHEET" | tr -d ' ')
if [ "$funded_count" -eq 0 ]; then
  die "FUNDED_ACCOUNTS is empty — a chain nobody can act on"
fi
if grep -qvE '^g1[0-9a-z]{38}=[1-9][0-9]*ugnot$' "$FUNDED_SHEET"; then
  die "FUNDED_ACCOUNTS has malformed entries (expected g1<38chars>=<digits>ugnot, one per entry): $(grep -vE '^g1[0-9a-z]{38}=[1-9][0-9]*ugnot$' "$FUNDED_SHEET" | tr '\n' ' ')"
fi
funded_dupes=$(cut -d= -f1 "$FUNDED_SHEET" | sort | uniq -d)
if [ -n "$funded_dupes" ]; then
  die "FUNDED_ACCOUNTS lists an address more than once: $funded_dupes"
fi
funded_total=$(sheet_total "$FUNDED_SHEET")
if [ "$funded_total" -ge "$FUNDED_TOTAL_CEILING" ]; then
  die "FUNDED_ACCOUNTS total $funded_total ugnot is at or above the $FUNDED_TOTAL_CEILING ceiling — too close to the int64 Coin limit for the fee-payer burn to be minted on top"
fi
print_substep "2.2" "Funded accounts: $funded_count, $funded_total ugnot (format + ceiling verified)"

# funded_amount <addr>  →  echoes the sheet amount for <addr>, or nothing.
funded_amount() {
  local line rc=0
  line=$(grep -m1 -- "^$1=" "$FUNDED_SHEET") || rc=$?
  if [ "$rc" -gt 1 ]; then
    die "grep failed looking up $1 in the funded sheet (exit $rc)"
  fi
  [ -n "$line" ] || return 0
  line="${line#*=}"
  printf '%s' "${line%%ugnot*}"
}

# ---- GovDAO T1 members must hold a genesis balance ----
# The bootstrap MsgRun seeds the T1 members, who then pay gas out of their own
# pocket for the chain's first proposals. onyx has a faucet, so this is not the
# lockout it is on mainnet — but the seeded member is also the validator
# operator and the only address that can `maketx run`, and a launch where the
# first governance action waits on a faucet drip is a launch with an
# unexplained gap. The addresses are read back out of the bootstrap source
# rather than repeated here, so this guard cannot drift from what gets seeded.
T1_ADDRS_FILE="$WORK_DIR/t1_members.txt"
# The sole genesis T1 member (aeddi), as on mainnet. Asserted rather than
# derived: a SetMember line dropped, duplicated or reshaped out of the grep's
# reach would otherwise shrink the set this guard covers without a word.
T1_EXPECTED_COUNT=1
BOOTSTRAP_GNO="$BOOTSTRAP_DIR/$(jq -r '.body_file' "$BOOTSTRAP_DIR/meta.json")"
if [ ! -f "$BOOTSTRAP_GNO" ]; then
  die "bootstrap body '$BOOTSTRAP_GNO' does not exist — check body_file in $BOOTSTRAP_DIR/meta.json"
fi
# Commented-out SetMember lines are stripped first: they carry the same shape
# but seed nobody. `|| true` lets a zero-match run reach the count check below
# instead of dying on grep's exit 1 under pipefail, unannounced.
grep -v '^[[:space:]]*//' "$BOOTSTRAP_GNO" |
  grep -oE 'memberstore\.T1, address\("g1[0-9a-z]{38}"\)' |
  grep -oE 'g1[0-9a-z]{38}' | sort -u >"$T1_ADDRS_FILE" || true
t1_count=$(wc -l <"$T1_ADDRS_FILE" | tr -d ' ')
if [ "$t1_count" -ne "$T1_EXPECTED_COUNT" ]; then
  die "$(printf '%s\n%s' \
    "expected $T1_EXPECTED_COUNT T1 members in $BOOTSTRAP_GNO, found $t1_count." \
    "A SetMember line was added, removed, duplicated or reshaped past this guard's grep — fix the file, or update T1_EXPECTED_COUNT if membership really changed.")"
fi

# The count above is over DISTINCT addresses, so it cannot see the same member
# seeded twice. MembersByTier.SetMember returns ErrMemberAlreadyExists on the
# second call and the bootstrap's must() turns it into a panic, which surfaces
# ~90 seconds later as a node crash during the measurement run instead of a
# sentence here. Counted per call, not per line.
t1_seed_calls=$(grep -v '^[[:space:]]*//' "$BOOTSTRAP_GNO" |
  grep -oE 'memberstore\.T1, address\("g1[0-9a-z]{38}"\)' | wc -l | tr -d ' ') || t1_seed_calls=0
if [ "$t1_seed_calls" -ne "$t1_count" ]; then
  die "$BOOTSTRAP_GNO makes $t1_seed_calls T1 SetMember calls for $t1_count distinct addresses — the repeated call fails with ErrMemberAlreadyExists and the bootstrap MsgRun panics at InitChain"
fi

# ---- The sole member's invitation points ----
# Every NewT1MemberRequest burns one point of the PROPOSER's balance when the
# proposal executes, and no realm path grants points afterwards, so the number
# typed into NewMember() at genesis is the permanent ceiling on how many
# members the launch member can ever seat. Nine, as on mainnet: onyx rehearses
# mainnet's governance, so its seed mirrors mainnet's.
T1_EXPECTED_INVITATION_POINTS=9
t1_points=$(grep -v '^[[:space:]]*//' "$BOOTSTRAP_GNO" |
  grep -oE 'memberstore\.T1, address\("g1[0-9a-z]{38}"\), memberstore\.NewMember\([0-9]+\)' |
  grep -oE 'NewMember\([0-9]+\)$' | grep -oE '[0-9]+') || t1_points=""
if [ "$t1_points" != "$T1_EXPECTED_INVITATION_POINTS" ]; then
  die "$(printf '%s\n%s' \
    "expected the sole T1 member in $BOOTSTRAP_GNO to be seeded with $T1_EXPECTED_INVITATION_POINTS invitation points, found '${t1_points:-none}'." \
    "That is mainnet's seed, and nothing on chain can top the balance up afterwards.")"
fi

t1_unfunded=""
while IFS= read -r t1_addr; do
  t1_amount=$(funded_amount "$t1_addr")
  if [ -z "$t1_amount" ]; then
    t1_unfunded="$t1_unfunded  $t1_addr — not in FUNDED_ACCOUNTS"$'\n'
  fi
done <"$T1_ADDRS_FILE"
if [ -n "$t1_unfunded" ]; then
  die "$(printf '%s\n%s%s' \
    "these GovDAO T1 members hold no genesis balance:" \
    "$t1_unfunded" \
    "add them to FUNDED_ACCOUNTS, or drop them from the bootstrap.")"
fi
print_substep "2.3" "GovDAO T1 members: $t1_count seeded with $t1_points invitation points, each holds a genesis balance"

# No founding-validator funding guard here, unlike mainnet's step 2.8: the
# validator's management actions (key rotation, profile edits, opt-out) are
# paid txs, but onyx has a faucet, and the operator (aeddi) is funded above
# anyway. The validator's signing address holds nothing at genesis.

# ---- NAMES_ADMIN must match the admin compiled into r/sys/names ----
# names.Enable is gated on an address hardcoded in the realm source, and
# NAMES_ADMIN is a copy of it. A stale copy surfaces ~90 seconds in as a
# "caller is not admin" panic during the measurement run; read the authority
# out of the tree instead. Step 6 checks the other half: that the tx's
# caller_override matches NAMES_ADMIN too.
NAMES_VERIFIER_GNO="$REPO_ROOT/examples/gno.land/r/sys/names/verifier.gno"
if [ ! -f "$NAMES_VERIFIER_GNO" ]; then
  die "cannot find $NAMES_VERIFIER_GNO — r/sys/names moved; re-point this check and re-confirm NAMES_ADMIN"
fi
names_admin_in_tree=$(grep -oE 'admin[[:space:]]*=[[:space:]]*address\("g1[0-9a-z]{38}"\)' "$NAMES_VERIFIER_GNO" |
  grep -oE 'g1[0-9a-z]{38}') || names_admin_in_tree=""
if [ -z "$names_admin_in_tree" ]; then
  die "no admin address found in $NAMES_VERIFIER_GNO — the realm's admin declaration changed shape; re-confirm NAMES_ADMIN by hand"
fi
if [ "$names_admin_in_tree" != "$NAMES_ADMIN" ]; then
  die "$(printf '%s\n%s\n%s' \
    "NAMES_ADMIN ($NAMES_ADMIN) is not the admin compiled into r/sys/names ($names_admin_in_tree)." \
    "names.Enable would be rejected at genesis: the caller_override patch only works for the address the realm actually trusts." \
    "Update NAMES_ADMIN and transactions/migration/names-enable/meta.json after confirming who that address belongs to.")"
fi
print_substep "2.4" "names admin matches r/sys/names in-tree: $names_admin_in_tree"

# ---- Inert policy: the four values have to agree with each other ----
# RUN_SUBMITTERS is assembled here, after 2.3 has read the seeded T1 members
# out of the bootstrap source, so the list cannot disagree with the membership
# the same build is about to seed.
RUN_SUBMITTERS=()
if [ "$RUN_SUBMITTERS_INCLUDE_GOVDAO_T1" = true ]; then
  while IFS= read -r t1_addr; do
    [ -n "$t1_addr" ] && RUN_SUBMITTERS+=("$t1_addr")
  done <"$T1_ADDRS_FILE"
fi
for extra_addr in "${RUN_SUBMITTERS_EXTRA[@]}"; do
  case " ${RUN_SUBMITTERS[*]} " in
  *" $extra_addr "*) die "RUN_SUBMITTERS_EXTRA lists $extra_addr, which is already a seeded GovDAO T1 member" ;;
  *) RUN_SUBMITTERS+=("$extra_addr") ;;
  esac
done

# Every one of these is a state the chain accepts and cannot repair on its
# own: Params.Validate checks address syntax and nothing else, and InitChain
# logs a note for an empty run_submitters but says nothing about an empty
# pkg_approvers. A genesis is forever, so they are build-time errors here.
case "$CODE_SUBMISSION_POLICY" in
permissionless | permissioned | inert) ;;
*) die "CODE_SUBMISSION_POLICY must be permissionless, permissioned or inert (got '$CODE_SUBMISSION_POLICY')" ;;
esac

# An approver list that is set under the wrong policy is as broken as an empty
# one under "inert" -- nothing would ever park, so the approvers would have
# nothing to enable and the key would be armed for no reason.
if [ "$CODE_SUBMISSION_POLICY" = inert ] && [ "${#PKG_APPROVERS[@]}" -eq 0 ]; then
  die "$(printf '%s\n%s' \
    "CODE_SUBMISSION_POLICY=inert with an empty PKG_APPROVERS: every post-genesis submission would park with nobody able to enable it, forever." \
    "Paste the gpao oracle address into PKG_APPROVERS.")"
fi
if [ "$CODE_SUBMISSION_POLICY" != inert ] && [ "${#PKG_APPROVERS[@]}" -gt 0 ]; then
  die "PKG_APPROVERS is set but CODE_SUBMISSION_POLICY is '$CODE_SUBMISSION_POLICY': nothing would ever park, so the approvers would have nothing to enable"
fi

for inert_addr in "${PKG_APPROVERS[@]}" "${RUN_SUBMITTERS[@]}"; do
  case "$inert_addr" in
  g1*) [ "${#inert_addr}" -eq 40 ] || die "malformed address in PKG_APPROVERS/RUN_SUBMITTERS: $inert_addr" ;;
  *) die "malformed address in PKG_APPROVERS/RUN_SUBMITTERS: $inert_addr" ;;
  esac
done

# An approver pays the gas for every MsgEnablePackage it sends. The faucet
# could top it up, but "the oracle approves nothing while every process on the
# host reports healthy" is the failure the inputs record warns about, so the
# balance is asserted here against the same sheet that funds it.
approver_unfunded=""
for inert_addr in "${PKG_APPROVERS[@]}"; do
  inert_amount=$(funded_amount "$inert_addr")
  if [ -z "$inert_amount" ]; then
    approver_unfunded="$approver_unfunded  $inert_addr — not in FUNDED_ACCOUNTS"$'\n'
  fi
done
if [ -n "$approver_unfunded" ]; then
  die "$(printf '%s\n%s%s' \
    "these package approvers hold no genesis balance:" \
    "$approver_unfunded" \
    "add them to FUNDED_ACCOUNTS, or the oracle can never enable anything.")"
fi

# RUN_SUBMITTERS armed must cover every seeded GovDAO T1 member.
#
# Proposal creation is MsgRun-only, so a T1 member missing from this list
# cannot propose -- and since amending the list is itself a proposal, a list
# that omits ALL of them is unamendable. Checked against the addresses read out
# of the bootstrap tx at 2.3, not a second copy typed here.
if [ "${#RUN_SUBMITTERS[@]}" -gt 0 ]; then
  run_missing=""
  while IFS= read -r t1_addr; do
    case " ${RUN_SUBMITTERS[*]} " in
    *" $t1_addr "*) ;;
    *) run_missing="$run_missing  $t1_addr"$'\n' ;;
    esac
  done <"$T1_ADDRS_FILE"
  if [ -n "$run_missing" ]; then
    die "$(printf '%s\n%s%s' \
      "RUN_SUBMITTERS is armed but omits these seeded GovDAO T1 members:" \
      "$run_missing" \
      "proposal creation is MsgRun-only, so they could not govern and the list could never be amended.")"
  fi
fi

if [ -z "$INERT_SUBMISSION_CHARGE" ] && [ -n "$INERT_CHARGE_COLLECTOR" ]; then
  die "INERT_CHARGE_COLLECTOR is set but INERT_SUBMISSION_CHARGE is empty: nothing would ever be collected"
fi
print_substep "2.5" "$(printf 'Code submission: policy=%s, approvers=%d, run_submitters=%d (gate %s), charge=%s' \
  "$CODE_SUBMISSION_POLICY" "${#PKG_APPROVERS[@]}" "${#RUN_SUBMITTERS[@]}" \
  "$([ "${#RUN_SUBMITTERS[@]}" -gt 0 ] && echo ARMED || echo OPEN)" "${INERT_SUBMISSION_CHARGE:-off}")"

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

# $pkg_count is what deplist resolved; this is what actually became a deploy.
# Three mechanisms drop a package between the two without a word: the staging
# copy at 4.3 uses `find -exec cp {} \;`, where a failed cp still leaves find
# exiting 0; ReadPkgListFromDir skips any directory whose gnomod.toml is
# missing (gnovm/pkg/packages/readpkglist.go); and GetNonIgnoredPkgs drops
# ignore-marked packages AND everything that depends on them
# (gnovm/pkg/packages/pkglist.go). A p/nt/* path that goes missing here
# cannot be added post-genesis — it is a relaunch.
addpkg_count=$(jq -r '[.tx.msg[] | select(.["@type"] == "/vm.m_addpkg")] | length' "$GENESIS_TXS_JSONL" |
  awk '{ s += $1 } END { print s + 0 }')
if [ "$addpkg_count" -ne "$pkg_count" ]; then
  die "$pkg_count packages resolved but $addpkg_count addpkg txs landed in the genesis — a package was dropped between staging and deploy (staging cp failure, missing gnomod.toml, or an ignore-marked dependency)"
fi
print_substep "4.8" "Reconciled: $addpkg_count addpkg txs for $pkg_count resolved packages"

# ---- Step 5: Add the bootstrap MsgRuns (transactions/base/)
# Seeds the sole GovDAO T1 member and locks AllowedDAOs, then registers the
# initial namespaces.

print_step_header 5 "$TOTAL_STEPS" "Add bootstrap MsgRuns (GovDAO seed + namespace preregistration)"

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

# ---- Namespace preregistration (transactions/base/users-preregister/) ----
# Registers the same initial names as mainnet in r/sys/users through the
# genesis-only r/sys/users/init.RegisterUser wrapper (the controller gate is
# skipped at height 0, so no authority survives genesis — see the .gno body's
# header). Ordered AFTER the addpkg stream like every genesis tx, so
# r/sys/users and r/sys/users/init exist when it runs.
USERS_PREREGISTER_DIR="$SCRIPT_DIR/transactions/base/users-preregister"
USERS_PREREGISTER_JSONL="$WORK_DIR/users_preregister_tx.jsonl"

print_substep "5.3" "Building AnnotatedTx from $USERS_PREREGISTER_DIR/..."
: >"$USERS_PREREGISTER_JSONL"
txn_dir_to_jsonl "$USERS_PREREGISTER_DIR" "$USERS_PREREGISTER_JSONL"

USERS_PREREGISTER_TX_FILE="$WORK_DIR/users_preregister_tx_stripped.jsonl"
jq -c 'del(.reason)' "$USERS_PREREGISTER_JSONL" >"$USERS_PREREGISTER_TX_FILE"

print_substep "5.4" "Adding namespace-preregistration tx to genesis..."
run "$GNOGENESIS_BIN" txs add sheets "$USERS_PREREGISTER_TX_FILE" --genesis-path "$GENESIS_FILE" 2>&1 | sed 's/^/    /'
cat "$USERS_PREREGISTER_TX_FILE" >>"$GENESIS_TXS_JSONL"

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
# hardfork mode), but the founding validator would have no operator-
# keyed management plane in r/sys/validators/v0.

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
    printf '%s,%s,%s,onyx founding validator (%s),cloud\n' \
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

# One Register per CSV row, or a founding validator silently launches with no
# operator-keyed valoper profile — no way to rotate its signing key or signal
# opt-out through r/sys/validators/v0, and the chain boots anyway.
valoper_tx_count=$(wc -l <"$VALOPER_SEED" | tr -d ' ')
if [ "$valoper_tx_count" -ne "${#INITIAL_VALSET[@]}" ]; then
  die "valoper-seed emitted $valoper_tx_count Register txs for ${#INITIAL_VALSET[@]} validators — a CSV row was rejected or dropped"
fi

print_substep "7.3" "Adding $valoper_tx_count valoper Register txs to genesis..."
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
# Spin up a temp node, pre-fund every creator/caller address with
# $INITIAL_BALANCE, let the genesis txs burn through fees, then query the
# remaining balances. The amount actually spent is what we credit each fee
# payer in the real genesis so its balance lands at zero post-genesis — the
# final state then holds ONLY the FUNDED_ACCOUNTS balances. The fee payers
# are the deployer (addpkgs, bootstrap, valoper Register fees), the names
# admin (names.Enable fee), and every gnomod.toml [addpkg] creator address in
# the package set (used as that package's addpkg creator in place of the
# deployer).
#
# Run twice for safety:
#   run 1: measure actual consumption with over-provisioned balances
#   run 2: verify the measured balances land everyone at the expected remainder
# If run 2 disagrees, something is non-deterministic and we abort.

# The code-submission vm params go into the shipping genesis BEFORE the
# measurement: step 8's params-parity assertion compares vm.params between
# the shipping and the temp-node genesis, and the temp side is patched with
# the same function. (Genesis replay is policy-exempt via IsGenesisReplay,
# so this changes no measured amount.) Step 9 reads the values back from
# the final artifact.
apply_code_submission_params "$GENESIS_FILE"

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
# creator/caller). Any other type names its fee payer differently and would
# end up unfunded — and an unfunded genesis signer is not an error: InitChain
# creates the account and MINTS it 10,000 GNOT (genesisSignerFunding,
# gno.land/pkg/gnoland/app.go), which lands in the supply counter. The chain
# would boot fine, with an unaccounted balance nobody chose and a supply the
# step 9 reconciliation cannot see, since it reads the genesis file rather
# than post-InitChain state.
unexpected_types=$(jq -r '.tx.msg[]["@type"]' "$GENESIS_TXS_JSONL" | sort -u |
  grep -vE '^/vm\.(m_addpkg|m_run|m_call)$' || true)
if [ -n "$unexpected_types" ]; then
  die "unexpected msg types in genesis txs: $unexpected_types"
fi
# Read the signer fields by path, not by pattern: package file bodies ride
# in these same lines, so a fixture containing the literal "caller":"g1..."
# would otherwise be funded as a phantom fee payer (it pays no fee, so the
# range guard in step 8.4 would abort the build on it).
jq -r '.tx.msg[] | (.creator // empty), (.caller // empty)' "$GENESIS_TXS_JSONL" |
  awk 'NF' |
  sort -u >"$BALANCES_TMP_CREATOR_ADDRESSES"
if grep -qvE '^g1[0-9a-z]{38}$' "$BALANCES_TMP_CREATOR_ADDRESSES"; then
  die "extracted a fee payer that is not a bech32 address: $(grep -vE '^g1[0-9a-z]{38}$' "$BALANCES_TMP_CREATOR_ADDRESSES" | tr '\n' ' ')"
fi
addr_count=$(wc -l <"$BALANCES_TMP_CREATOR_ADDRESSES" | tr -d ' ')
# Dozens of txs cannot have zero signers: an empty extraction means the signer
# field moved, and every downstream count check would compare zero against zero.
if [ "$addr_count" -eq 0 ]; then
  die "no creator/caller extracted from $GENESIS_TXS_JSONL — the signer field shape changed"
fi
print_substep "8.2" "Found $addr_count unique creator/caller addresses"

# Overlap rule (the final balance sheet keeps one entry per address, last
# write wins, so every overlap must be resolved explicitly): a fee payer MAY
# also be a FUNDED_ACCOUNTS entry — its final entry becomes funded amount +
# measured burn, so it lands at exactly its funded amount once the genesis
# txs execute. Resolved during the measure run below, so the verify run
# replays the exact entries that ship.

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

  # The temp node runs with the shipped bank params — an empty lock on onyx —
  # written explicitly rather than left to the generator, so the measurement
  # and the shipping genesis cannot disagree on them (step 9 writes the same
  # value and asserts it).
  temp_rd_json=$(printf '%s\n' "${RESTRICTED_DENOMS[@]}" | jq -R -s -c 'split("\n") | map(select(length > 0))')
  jq --argjson rd "$temp_rd_json" '.app_state.bank.params.restricted_denoms = $rd' \
    "$BALANCES_TMP_GENESIS" >"$BALANCES_TMP_GENESIS.locked"
  mv "$BALANCES_TMP_GENESIS.locked" "$BALANCES_TMP_GENESIS"
  if [ "$(jq -c '.app_state.bank.params.restricted_denoms' "$BALANCES_TMP_GENESIS")" != "$temp_rd_json" ]; then
    die "temp-node genesis did not take restricted_denoms=$temp_rd_json"
  fi
  # Same code-submission params as the shipping genesis. The measurement
  # replays the genesis txs, which are exempt from the policy via
  # IsGenesisReplay, so this changes no measured amount -- it is here because
  # the parity assertion below compares vm.params.
  apply_code_submission_params "$BALANCES_TMP_GENESIS"

  # The measured burns are dominated by storage deposits priced by the
  # chain's vm/auth params. This genesis is regenerated rather than copied
  # (a future GENESIS_TIME would stall the temp node), so assert its
  # fee-governing params equal the shipping genesis AS IT STANDS NOW —
  # otherwise every measured amount is wrong by the parameter ratio and run 2
  # would agree with it.
  params_parity_lhs=$(jq -cS '[.app_state.vm.params, .app_state.auth.params]' "$BALANCES_TMP_GENESIS")
  if [ "$params_parity_lhs" = "[null,null]" ]; then
    die "temp-node genesis carries no vm/auth params (app_state moved?) — the parity check below would compare nothing"
  fi
  if [ "$params_parity_lhs" != \
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
    # case instead of `echo | grep -q`, and no `| head -1` after the sed:
    # early-exit consumers are the SIGPIPE-under-pipefail shape that
    # account_in_state was rewritten to avoid. gnokey prints exactly one
    # `data:` line; if that ever changes, the multi-line payload fails the
    # exact-match parses below loudly instead of being truncated silently.
    case "$out" in
    'data:'* | *$'\n''data:'*)
      local payload
      payload=$(echo "$out" | sed -n 's/^data: //p')
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
      ;;
    esac
    sleep 1
    retry=$((retry + 1))
  done
  echo "ERROR: Could not query balance for $addr after ${NODE_TIMEOUT}s." >&2
  exit 1
}

start_temp_node "run 1: measure gas costs"
print_substep "8.4" "Querying remaining balances..."
rm -f "$BALANCES_TMP_FILE"
EXPECTED_REMAINDERS="$BALANCES_TMP_DIR/expected-remainders.txt"
OVERLAP_ADDRS="$BALANCES_TMP_DIR/funded-overlap-addrs.txt"
: >"$EXPECTED_REMAINDERS"
: >"$OVERLAP_ADDRS"
merged_funded_total=0
while IFS= read -r addr; do
  remaining=$(query_balance "$addr")
  # A fee payer funded with the float and charged at least one fee must
  # land strictly between 0 and INITIAL_BALANCE.
  if [ "$remaining" -eq 0 ] || [ "$remaining" -ge "$INITIAL_BALANCE" ]; then
    die "fee payer $addr reads $remaining ugnot remaining in the measure run (float: $INITIAL_BALANCE) — node state not readable or fees not charged"
  fi
  final=$((INITIAL_BALANCE - remaining))
  # A fee payer that is also a FUNDED_ACCOUNTS entry gets ONE sheet entry of
  # funded amount + burn: post-genesis it lands at exactly its funded amount
  # instead of zero. The merge happens HERE, before run 2, so the verify run
  # replays the exact entries that ship and checks each expected remainder.
  fp_funded=$(funded_amount "$addr")
  if [ -n "$fp_funded" ]; then
    merged_funded_total=$((merged_funded_total + fp_funded))
    printf "    %s = %s ugnot (+ %s funded)\n" "$addr" "$final" "$fp_funded"
    echo "${addr}=$((final + fp_funded))ugnot" >>"$BALANCES_TMP_FILE"
    echo "${addr} ${fp_funded}" >>"$EXPECTED_REMAINDERS"
    echo "$addr" >>"$OVERLAP_ADDRS"
  else
    printf "    %s = %s ugnot\n" "$addr" "$final"
    echo "${addr}=${final}ugnot" >>"$BALANCES_TMP_FILE"
    echo "${addr} 0" >>"$EXPECTED_REMAINDERS"
  fi
done <"$BALANCES_TMP_CREATOR_ADDRESSES"
stop_temp_node

start_temp_node "run 2: verify expected remainders"
print_substep "8.5" "Verifying fee payers land at their expected remainders..."
all_expected=true
verified_count=0
while IFS=' ' read -r addr expected; do
  verified_count=$((verified_count + 1))
  remaining=$(query_balance "$addr")
  if [ "$remaining" -ne "$expected" ]; then
    printf "    FAIL: %s has %sugnot remaining (expected %s)\n" "$addr" "$remaining" "$expected"
    all_expected=false
  else
    printf "    ok: %s (%s)\n" "$addr" "$expected"
  fi
done <"$EXPECTED_REMAINDERS"
stop_temp_node

if [ "$all_expected" != true ]; then
  die "Some fee payers did not land at their expected remainder. Check $BALANCES_TMP_FILE."
fi
# An empty expected-remainders file would sail through the loop above and
# report verified costs having checked nothing.
if [ "$verified_count" -ne "$addr_count" ]; then
  die "verified $verified_count fee payers but $addr_count were funded — $EXPECTED_REMAINDERS is short"
fi
print_substep "8.6" "All fee payers land at their expected remainders — costs verified"
cp "$BALANCES_TMP_FILE" "$DEPLOYER_BALANCES"
deployer_rows=$(wc -l <"$DEPLOYER_BALANCES" | tr -d ' ')
if [ "$deployer_rows" -ne "$addr_count" ]; then
  die "the measured fee-payer sheet has $deployer_rows rows for $addr_count fee payers — a row was dropped or duplicated between the measurement and the shipped sheet"
fi
# This small sheet is where the MEASURED numbers live; locking it localises a
# future genesis.json mismatch to "the burn changed" instead of "the bytes
# changed somewhere".
verify_checksum "$DEPLOYER_BALANCES"

# Strip the merged addresses from the funded sheet, then assert the fee-payer
# and funded sheets are disjoint — one entry per address in the final
# genesis, and last-write-wins would silently drop one side of any overlap.
FUNDED_STRIPPED="$WORK_DIR/funded_stripped.txt"
merged_count=$(wc -l <"$OVERLAP_ADDRS" | tr -d ' ')
if [ -s "$OVERLAP_ADDRS" ]; then
  sed 's/^/^/; s/$/=/' "$OVERLAP_ADDRS" >"$OVERLAP_ADDRS.pat"
  rc=0
  grep -v -f "$OVERLAP_ADDRS.pat" "$FUNDED_SHEET" >"$FUNDED_STRIPPED" || rc=$?
  if [ "$rc" -gt 1 ]; then
    die "grep failed stripping merged addresses from the funded sheet (exit $rc)"
  fi
else
  cp "$FUNDED_SHEET" "$FUNDED_STRIPPED"
fi
sheet_overlap=$(comm -12 <(cut -d= -f1 "$FUNDED_STRIPPED" | sort) <(cut -d= -f1 "$DEPLOYER_BALANCES" | sort))
if [ -n "$sheet_overlap" ]; then
  die "fee-payer and funded sheets are not disjoint after the merge: $sheet_overlap"
fi

# ---- Step 9: Add validators + balances, verify, move into place

print_step_header 9 "$TOTAL_STEPS" "Add validators + balances, verify genesis"

print_substep "9.1" "Adding the initial validator set..."
for validator in "${INITIAL_VALSET[@]}"; do
  read -r name power address pub_key <<<"$validator"
  printf "    %s (power=%s, %s)\n" "$name" "$power" "$address"
  run "$GNOGENESIS_BIN" validator add -name "$name" -power "$power" -address "$address" -pub-key "$pub_key" --genesis-path "$GENESIS_FILE"
done

# Fee payers (exact burn, or funded amount + burn for the funded ones) + the
# remaining funded accounts, one sheet. One entry per address (last write
# wins), so every overlap was resolved by step 8.
FULL_BALANCES_FILE="$WORK_DIR/balances.txt"
cat "$DEPLOYER_BALANCES" "$FUNDED_STRIPPED" >"$FULL_BALANCES_FILE"
balance_count=$(wc -l <"$FULL_BALANCES_FILE" | tr -d ' ')
funded_stripped_count=$(wc -l <"$FUNDED_STRIPPED" | tr -d ' ')
print_substep "9.2" "Adding $balance_count balances ($addr_count fee payers, $merged_count of them funded, + $funded_stripped_count funded accounts)..."
run "$GNOGENESIS_BIN" balances add -balance-sheet "$FULL_BALANCES_FILE" --genesis-path "$GENESIS_FILE" >/dev/null

# ---- Bank params: transfers open ----
# Written explicitly (the same value the temp nodes ran with) and read back:
# a genesis that silently carried a lock would strand the faucets at block 1.
print_substep "9.3" "Applying the bank params (transfers open)..."
restricted_json=$(printf '%s\n' "${RESTRICTED_DENOMS[@]}" | jq -R -s -c 'split("\n") | map(select(length > 0))')
jq --argjson rd "$restricted_json" '.app_state.bank.params.restricted_denoms = $rd' \
  "$GENESIS_FILE" >"$GENESIS_FILE.bank"
mv "$GENESIS_FILE.bank" "$GENESIS_FILE"
applied_rd=$(jq -r '.app_state.bank.params.restricted_denoms | length' "$GENESIS_FILE")
applied_unres=$(jq -r '.app_state.auth.params.unrestricted_addrs | length // 0' "$GENESIS_FILE")
if [ "$applied_rd" -ne "${#RESTRICTED_DENOMS[@]}" ]; then
  die "restricted_denoms did not apply: wanted ${#RESTRICTED_DENOMS[@]} entries, genesis has $applied_rd"
fi
if [ "${#RESTRICTED_DENOMS[@]}" -eq 0 ] && [ "$applied_unres" -ne 0 ]; then
  die "the genesis carries $applied_unres unrestricted_addrs with no transfer lock — an exemption list for a lock that does not exist"
fi

# ---- Code-submission policy (inert) ----
# Applied BEFORE step 8 (the params-parity assertion there requires the
# shipping and measurement genesis to carry identical vm.params); read back
# here from the artifact that ships.
applied_policy=$(jq -r '.app_state.vm.params.code_submission_policy' "$GENESIS_FILE")
applied_approvers=$(jq -r '.app_state.vm.params.pkg_approvers | length // 0' "$GENESIS_FILE")
applied_runners=$(jq -r '.app_state.vm.params.run_submitters | length // 0' "$GENESIS_FILE")
print_substep "9.4" "$(printf 'Transfers: %s; code submission: policy=%s, %d approver(s), MsgRun gate %s (%d)' \
  "$([ "$applied_rd" -eq 0 ] && echo OPEN || echo "LOCKED on $applied_rd denom(s)")" \
  "$applied_policy" "$applied_approvers" \
  "$([ "$applied_runners" -gt 0 ] && echo ARMED || echo OPEN)" "$applied_runners")"

# ---- Reconcile account count and total supply ----
# Everything above counts lines in files the script itself wrote. This reads
# the account count and the supply back out of the genesis that ships and ties
# both to the two sources they must come from: FUNDED_ACCOUNTS as typed, and
# the burn measured on the temp node. A truncated `balances add`, a sheet
# added twice or a mis-stripped merge row all break one of these identities.
#
# The two launch parameters nothing downstream would notice: `gnogenesis
# verify` has no opinion on either, and the temp nodes deliberately run with
# their own generated genesis, so a `generate` call that silently dropped one
# would ship a chain nobody can join.
genesis_chain_id=$(jq -r '.chain_id' "$GENESIS_FILE")
if [ "$genesis_chain_id" != "$CHAIN_ID" ]; then
  die "shipped genesis carries chain_id '$genesis_chain_id', expected '$CHAIN_ID'"
fi
genesis_time_shipped=$(jq -r '.genesis_time' "$GENESIS_FILE")
genesis_time_want=$(date -u -r "$GENESIS_TIME" +%Y-%m-%dT%H:%M:%SZ 2>/dev/null ||
  date -u -d "@$GENESIS_TIME" +%Y-%m-%dT%H:%M:%SZ)
if [ "$genesis_time_shipped" != "$genesis_time_want" ]; then
  die "shipped genesis carries genesis_time '$genesis_time_shipped', expected '$genesis_time_want' (GENESIS_TIME=$GENESIS_TIME)"
fi

genesis_accounts=$(jq -r '.app_state.balances | length' "$GENESIS_FILE")
genesis_supply=$(jq -r '.app_state.balances[]' "$GENESIS_FILE" | sheet_total /dev/stdin)

fee_payer_total=$(sheet_total "$DEPLOYER_BALANCES")
burn_total=$((fee_payer_total - merged_funded_total))

if [ "$genesis_accounts" -ne "$balance_count" ]; then
  die "genesis holds $genesis_accounts balance entries, expected $balance_count"
fi
if [ "$funded_stripped_count" -ne $((funded_count - merged_count)) ]; then
  die "the stripped funded sheet has $funded_stripped_count rows, expected $((funded_count - merged_count)) ($funded_count typed less $merged_count merged into fee-payer entries)"
fi
expected_supply=$((funded_total + burn_total))
if [ "$genesis_supply" -ne "$expected_supply" ]; then
  die "$(printf '%s\n%s\n%s' \
    "genesis supply is $genesis_supply ugnot, expected $expected_supply:" \
    "  funded accounts (typed)    $funded_total" \
    "  fee-payer burn (measured)  $burn_total")"
fi
print_substep "9.5" "$(printf 'Reconciled: %s accounts, %s ugnot = %s funded + %s burn' \
  "$genesis_accounts" "$genesis_supply" "$funded_total" "$burn_total")"

print_substep "9.6" "Running gnogenesis verify..."
# -skip-signature-check: the names.Enable tx carries a post-sign caller
# patch and the valoper Register tx carries a placeholder signature, so
# per-tx signature verification cannot pass by design (nodes accept both
# under --skip-genesis-sig-verification). The remaining scope is shallow —
# amino decode, GenesisDoc/params sanity, stateless per-tx and per-balance
# format checks. Execution and fee-payer funding are proven by step 8's
# temp-node runs, not by this step.
run "$GNOGENESIS_BIN" verify -genesis-path "$GENESIS_FILE" -skip-signature-check

# Verify before moving: a mismatch must not clobber the previously-good
# (gitignored, so invisible to git status) genesis.json at the root.
verify_checksum "$GENESIS_FILE" genesis.json
print_substep "9.7" "Moving $GENESIS_FILE -> $FINAL_GENESIS"
mv "$GENESIS_FILE" "$FINAL_GENESIS"

# ---- Summary

PIPELINE_END_TS=$(date +%s)
PIPELINE_DURATION=$((PIPELINE_END_TS - PIPELINE_START_TS))
FINAL_SHA=$(sha256_of "$FINAL_GENESIS")
FINAL_BYTES=$(file_size "$FINAL_GENESIS")

printf '\n### onyx build complete: genesis.json (%s, sha256=%s) ###\n' \
  "$(format_size "$FINAL_BYTES")" "$FINAL_SHA"
printf '    total pipeline time: %s\n' "$(format_duration "$PIPELINE_DURATION")"

# What produced these bytes. packages.gen.txt records package PATHS, not
# content, so the source tree is the rest of the input: two runs from
# different commits — or from a dirty tree — legitimately differ, and without
# this a future checksum mismatch is unexplainable. Both inputs count:
# examples/ is what gets deployed, and this folder is what assembles it.
SOURCE_COMMIT=$(git -C "$REPO_ROOT" rev-parse HEAD 2>/dev/null || echo "unknown")
SOURCE_DIRTY=$(git -C "$REPO_ROOT" status --porcelain -- examples/ "$DEPLOY_DIR" 2>/dev/null | wc -l | tr -d ' ')
printf '    source commit:       %s%s\n' "$SOURCE_COMMIT" \
  "$([ "$SOURCE_DIRTY" -gt 0 ] && printf ' + %s UNCOMMITTED path(s) in examples/ or this folder' "$SOURCE_DIRTY")"

# The checksum manifest is the only thing standing between "the bytes I
# reviewed" and "the bytes the validator boots", so say plainly how much of
# this build it covered.
printf '    artifacts locked:    %s of %s\n' "$CHECKSUMS_LOCKED" "$((CHECKSUMS_LOCKED + CHECKSUMS_UNLOCKED))"
if [ "$CHECKSUMS_UNLOCKED" -gt 0 ]; then
  printf '\n    %s artifact(s) NOT locked — this build is not reproducible-verified.\n' "$CHECKSUMS_UNLOCKED"
  printf '    Paste into CHECKSUMS_DATA once every launch value is final:\n\n'
  printf '%s' "$CHECKSUMS_UNLOCKED_LINES" | sed 's/^/      /'
  printf '\n'
fi
