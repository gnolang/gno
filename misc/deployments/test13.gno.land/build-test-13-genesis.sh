#!/usr/bin/env bash
# build-test-13-genesis.sh — produces the test-13 BASE genesis (no historical
# replay yet). The output is consumed by apply-test-13-replay.sh, which
# appends gnoland1's historical txs + the T1 rotation migration.
#
# This script is a near-clone of misc/deployments/gnoland1/gen-genesis.sh,
# with four deltas that make the result a "hardfork-ready" base instead
# of gnoland1's own genesis:
#
#   1. FILTERED_PACKAGES adds the four packages that gnoland1 doesn't ship
#      (p/onbloc/{uint256,int256,json}, r/sys/validators/v3) and one realm
#      (r/demo/defi/grc20reg) we want available post-fork.
#   2. INITIAL_VALSET is the test-13 valset rather than gnoland1's
#      launch 7. Power 10 each — chosen for visibility in tooling that
#      displays the ratio of votes; consensus only cares about
#      relative weight.
#   3. govdao_prop1_test13.gno replaces gnoland1's govdao_prop1.gno. It
#      drops the v2 valset seed (we set the consensus valset directly via
#      GenesisDoc.Validators), adds 10 faucet addresses to the
#      ProposeAddUnrestrictedAcctsRequest call so we can transact under
#      the restricted-denom regime without manual unrestrict txs, and
#      wires the rotate-pkg realm (delta 4) into AllowedDAOs at lock time.
#   4. rotate-pkg/rotate.gno is a single-use realm addpkg'd alongside the
#      filtered set. A phase-2 MsgCall to its Rotate() function swaps the
#      sole T1 from manfred (gnoland1 inherit) to the test-13 T1, then
#      Rotate self-ejects from AllowedDAOs. See rotate-pkg/rotate.gno for
#      why proposal-flow rotation is unworkable across migration MsgRuns.
#
# Output: ./out/base-genesis.json (gitignored).
#
# Usage:
#   ./build-test-13-genesis.sh              # full build
#   ./build-test-13-genesis.sh --debug      # show every command being run
#   ./build-test-13-genesis.sh --no-install # reuse previously built binaries
#   ./build-test-13-genesis.sh --txs-only   # stop after generating txs (skip balance calc)
set -eo pipefail

# =============================================================================
# Launch parameters — review before each genesis generation.
# =============================================================================

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
  "aeddi-1 10 g1s2ht24e85qq3t66gc9sgdvk5kzc38yy68aaqvr gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqfr74tgql2cvzadga2uts62v3f8a5dx66dauaq6sphg3ynuhgl286cce2mn"
  "gfanton-1 10 g1kq98592x93lu29smp6lufjvyj9fqentg93x69f gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqtg9d2yxyp07etaqtdx9ryrsreun6h764ctm6faaa2lgq7e2jd6ecderqe8"
)

# Faucet balances. Each gets $FAUCET_BALANCE ugnot at genesis. Addresses
# are pasted from `gnokey list` output of an off-tree keybase (mnemonics
# are NOT in this repo). The faucet addresses MUST also appear in
# govdao_prop1_test13.gno's ProposeAddUnrestrictedAcctsRequest call so
# they can transfer ugnot under the locked-bank regime; the two lists are
# kept in sync by hand for now (10 entries, low maintenance churn).
FAUCET_BALANCE=100000000000 # 100M ugnot per faucet
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

# Chain parameters.
CHAIN_ID=test-13
GENESIS_TIME=1773651600 # Monday, March 16th 2026 10:00 GMT+0100 (CET) — same as gnoland1

# Airdrop balances — reused as-is from gnoland1. The fork promise is to
# preserve gnoland1's balance state, so the airdrop snapshot lands here
# verbatim.
BALANCES_GZ_URL="https://github.com/gnolang/independence-day/raw/9dec38a4a72c9e84db7e78ae010370de250f2d64/mkgenesis/balances.txt.gz"

# =============================================================================
# Internal — everything below is glue, you shouldn't need to change it.
# =============================================================================

# Deployer key mnemonic (deterministic — used only to sign genesis-mode txs).
# Same as gnoland1 so the deployer address is reproducible across both chains.
DEPLOYER_MNEMONIC="anchor hurt name seed oak spread anchor filter lesson shaft wasp home improve text behind toe segment lamp turn marriage female royal twice wealth"

# ---- Flags

STOP_AFTER_TXS_EXPORT=false
DEBUG=false
NO_INSTALL=false
for arg in "$@"; do
  case "$arg" in
  --txs-only) STOP_AFTER_TXS_EXPORT=true ;;
  --debug) DEBUG=true ;;
  --no-install) NO_INSTALL=true ;;
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

# Clean up temp node on exit.
NODE_PID=""
cleanup() { [ -n "$NODE_PID" ] && kill "$NODE_PID" 2>/dev/null || true; }
trap cleanup EXIT

# ---- Derived paths

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENESIS_FILE="$SCRIPT_DIR/out/base-genesis.json"
WORK_DIR="$SCRIPT_DIR/genesis-work"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
EXAMPLES_DIR="$REPO_ROOT/examples"
GNO_CMD="$REPO_ROOT/gnovm/cmd/gno"
GNOKEY_CMD="$REPO_ROOT/gno.land/cmd/gnokey"
GNOLAND_CMD="$REPO_ROOT/gno.land/cmd/gnoland"
GNOGENESIS_CMD="$REPO_ROOT/contribs/gnogenesis"
WORK_DIR_BIN="$WORK_DIR/bin"
GNO_BIN="$WORK_DIR_BIN/gno"
GNOKEY_BIN="$WORK_DIR_BIN/gnokey"
GNOLAND_BIN="$WORK_DIR_BIN/gnoland"
GNOGENESIS_BIN="$WORK_DIR_BIN/gnogenesis"

mkdir -p "$SCRIPT_DIR/out"

# Clean up previous work directory (preserve bin/ when --no-install).
if [ "$NO_INSTALL" = true ]; then
  find "$WORK_DIR" -mindepth 1 -maxdepth 1 ! -name bin -exec rm -rf {} + 2>/dev/null || true
else
  rm -rf "$WORK_DIR"
fi

# ---- 1. Build binaries from source

if [ "$NO_INSTALL" = true ]; then
  printf "\n=== Step 1/8: Skipping build (--no-install) ===\n"
  for bin in "$GNO_BIN" "$GNOKEY_BIN" "$GNOLAND_BIN" "$GNOGENESIS_BIN"; do
    if [ ! -x "$bin" ]; then
      echo "ERROR: --no-install but $bin not found. Run without --no-install first."
      exit 1
    fi
  done
else
  printf "\n=== Step 1/8: Building binaries ===\n"
  mkdir -p "$WORK_DIR_BIN"

  printf "  gno...        "
  run go build -C "$GNO_CMD" -o "$GNO_BIN" .
  printf "ok\n"

  printf "  gnokey...     "
  run go build -C "$GNOKEY_CMD" -o "$GNOKEY_BIN" .
  printf "ok\n"

  printf "  gnoland...    "
  run go build -C "$GNOLAND_CMD" -o "$GNOLAND_BIN" .
  printf "ok\n"

  printf "  gnogenesis... "
  run go build -C "$GNOGENESIS_CMD" -o "$GNOGENESIS_BIN" .
  printf "ok\n"
fi

# ---- 2. Generate filtered examples genesis txs.

printf "\n=== Step 2/8: Generating addpkg txs ===\n"

printf "  Resolving dependencies...\n"
pkg_dirs=$(cd "$EXAMPLES_DIR" && "$GNO_BIN" tool deplist -test-dep "${FILTERED_PACKAGES[@]}")
pkg_count=$(echo "$pkg_dirs" | wc -l | tr -d ' ')
printf "  Resolved %s packages in topological order\n" "$pkg_count"

# Save resolved package list for inspection.
{
  echo "# Generated by build-test-13-genesis.sh — do not edit."
  echo "$pkg_dirs" | sed "s|$EXAMPLES_DIR/||g"
} >"$WORK_DIR/packages.gen.txt"

printf "  Copying packages to staging...\n"
WORK_DIR_EXAMPLES="$WORK_DIR/examples"
mkdir -p "$WORK_DIR_EXAMPLES"
while IFS= read -r dir; do
  [[ -z "$dir" ]] && continue
  rel="${dir#$EXAMPLES_DIR/}"
  dest="$WORK_DIR_EXAMPLES/$rel"
  mkdir -p "$dest"
  find "$dir" -maxdepth 1 -type f -exec cp {} "$dest/" \;
  [[ -d "$dir/filetests" ]] && cp -r "$dir/filetests" "$dest/filetests"
done <<<"$pkg_dirs"

# Stage the test-13 single-use rotation realm. Lives outside examples/
# because it ships only for this hardfork — the bootstrap MsgRun puts it
# in AllowedDAOs at lock time, a phase-2 MsgCall to Rotate() swaps the
# sole T1 from manfred to the test-13 T1, then Rotate self-ejects.
# LoadPackagesFromDir topo-sorts on the gnomod.toml dependency graph, so
# rotate's addpkg lands after gov/dao + memberstore even though it's
# copied here separately. See rotate-pkg/rotate.gno for the design.
ROTATE_PKG_SRC="$SCRIPT_DIR/rotate-pkg"
ROTATE_PKG_DEST="$WORK_DIR_EXAMPLES/gno.land/r/test13/rotate"
printf "  Staging single-use rotation realm at %s...\n" "${ROTATE_PKG_DEST#$WORK_DIR_EXAMPLES/}"
mkdir -p "$ROTATE_PKG_DEST"
find "$ROTATE_PKG_SRC" -maxdepth 1 -type f -exec cp {} "$ROTATE_PKG_DEST/" \;

printf "  Creating deployer key...\n"
WORK_DIR_GNOKEY_HOME="$WORK_DIR/gnokey-home"
WORK_DIR_GENESIS="$WORK_DIR/genesis.json"
WORK_DIR_GENESIS_TXS="$WORK_DIR/genesis_txs.jsonl"
printf '%s\n\n' "$DEPLOYER_MNEMONIC" | run "$GNOKEY_BIN" add --recover GenesisDeployer --home "$WORK_DIR_GNOKEY_HOME" --insecure-password-stdin 2>&1 | sed 's/^/    /'

printf "  Generating empty genesis...\n"
run "$GNOGENESIS_BIN" generate -chain-id "$CHAIN_ID" -genesis-time "$GENESIS_TIME" --output-path "$WORK_DIR_GENESIS" 2>&1 | sed 's/^/    /'

printf "  Adding %s packages to genesis...\n" "$pkg_count"
echo "" | run "$GNOGENESIS_BIN" txs add packages "$WORK_DIR_EXAMPLES" -gno-home "$WORK_DIR_GNOKEY_HOME" -key-name GenesisDeployer --genesis-path "$WORK_DIR_GENESIS" --insecure-password-stdin 2>&1 | sed 's/^/    /'

printf "  Exporting txs...\n"
run "$GNOGENESIS_BIN" txs export "$WORK_DIR_GENESIS_TXS" --genesis-path "$WORK_DIR_GENESIS" 2>&1 | sed 's/^/    /'

# ---- 3. Generate setup transaction (govdao_prop1_test13.gno)

printf "\n=== Step 3/8: Generating MsgRun setup tx (govdao_prop1_test13.gno) ===\n"

SETUP_FILE="$SCRIPT_DIR/govdao_prop1_test13.gno"

printf "  Generating MsgRun tx from %s...\n" "$(basename "$SETUP_FILE")"
SETUP_TX="$WORK_DIR/genesis_setup_tx.json"
SETUP_TX_FILE="$WORK_DIR/genesis_setup_tx.jsonl"
run "$GNOKEY_BIN" maketx run \
  --gas-wanted 100000000 \
  --gas-fee 1ugnot \
  --chainid "$CHAIN_ID" \
  --home "$WORK_DIR_GNOKEY_HOME" \
  --broadcast=false \
  --insecure-password-stdin \
  GenesisDeployer \
  "$SETUP_FILE" >"$SETUP_TX" <<<""

printf "  Signing tx...\n"
echo "" | run "$GNOKEY_BIN" sign \
  --tx-path "$SETUP_TX" \
  --chainid "$CHAIN_ID" \
  --account-number 0 \
  --account-sequence 0 \
  --home "$WORK_DIR_GNOKEY_HOME" \
  --insecure-password-stdin \
  GenesisDeployer
jq -c '{tx: .}' <"$SETUP_TX" >"$SETUP_TX_FILE"

printf "  Adding setup tx to genesis...\n"
run "$GNOGENESIS_BIN" txs add sheets "$SETUP_TX_FILE" --genesis-path "$WORK_DIR_GENESIS" 2>&1 | sed 's/^/    /'
cat "$SETUP_TX_FILE" >>"$WORK_DIR_GENESIS_TXS"

tx_count=$(wc -l <"$WORK_DIR_GENESIS_TXS" | tr -d ' ')
printf "  Total txs so far: %s\n" "$tx_count"

if [ "$STOP_AFTER_TXS_EXPORT" = true ]; then
  cp "$WORK_DIR/packages.gen.txt" "$SCRIPT_DIR/packages.gen.txt"
  cp "$WORK_DIR_GENESIS_TXS" "$SCRIPT_DIR/genesis_txs.jsonl"
  printf "\n=== Done (--txs-only) ===\n"
  printf "  %s txs exported\n" "$tx_count"
  exit 0
fi

# ---- 4. Calculate deployer balances
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

printf "\n=== Step 4/8: Calculating deployer balances ===\n"

WORK_DIR_DEPLOYER_BALANCES="$WORK_DIR/deployers_balances.txt"
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

printf "  Extracting creator addresses...\n"
grep -oE '"(creator|caller)":"[^"]*"' "$WORK_DIR_GENESIS_TXS" |
  sed 's/"creator":"//;s/"caller":"//;s/"//g' |
  sort -u >"$BALANCES_TMP_CREATOR_ADDRESSES"
addr_count=$(wc -l <"$BALANCES_TMP_CREATOR_ADDRESSES" | tr -d ' ')
printf "  Found %s unique creator/caller addresses\n" "$addr_count"

printf "  Generating over-provisioned balances...\n"
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
      echo "ERROR: Node stopped unexpectedly. Last log lines:"
      tail -20 "$BALANCES_TMP_GNOLAND_LOG"
      exit 1
    fi
    if curl -sf "http://$NODE_RPC_ADDR/status" >/dev/null 2>&1; then
      printf "  Node ready (%ss)\n" "$elapsed"
      return
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  kill "$NODE_PID" 2>/dev/null || true
  echo "ERROR: Node did not start within ${NODE_TIMEOUT}s. Last log lines:"
  tail -20 "$BALANCES_TMP_GNOLAND_LOG"
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
printf "  Querying remaining balances...\n"
rm -f "$BALANCES_TMP_FILE"
while IFS= read -r addr; do
  remaining=$(query_balance "$addr")
  final=$((INITIAL_BALANCE - remaining))
  printf "    %s = %s ugnot\n" "$addr" "$final"
  echo "${addr}=${final}ugnot" >>"$BALANCES_TMP_FILE"
done <"$BALANCES_TMP_CREATOR_ADDRESSES"
stop_temp_node

start_temp_node "run 2: verify zero balances"
printf "  Verifying all balances are zero...\n"
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
  echo "ERROR: Some balances are not zero after replay. Check $BALANCES_TMP_FILE."
  exit 1
fi
printf "  All balances zero — deployer costs verified\n"
cp "$BALANCES_TMP_FILE" "$WORK_DIR_DEPLOYER_BALANCES"

# ---- 5. Add the initial validator set to the genesis file
# Done before balances — gnogenesis is O(filesize) per call, so adding
# validators while the file is still small saves ~6 minutes.

printf "\n=== Step 5/8: Adding validators ===\n"

for validator in "${INITIAL_VALSET[@]}"; do
  read -r name power address pub_key <<<"$validator"
  printf "  %s (power=%s, %s)\n" "$name" "$power" "$address"
  run "$GNOGENESIS_BIN" validator add -name "$name" -power "$power" -address "$address" -pub-key "$pub_key" --genesis-path "$WORK_DIR_GENESIS"
done

# ---- 6. Add gnoland1-canonical balances (deployer + airdrop)
# Faucets are deliberately skipped here and appended in step 7 instead;
# see step 7 for the SignerInfo-ordering rationale.

printf "\n=== Step 6/8: Adding deployer + airdrop balances ===\n"

printf "  Adding deployer balances to genesis...\n"
run "$GNOGENESIS_BIN" balances add -balance-sheet "$WORK_DIR_DEPLOYER_BALANCES" --genesis-path "$WORK_DIR_GENESIS" >/dev/null

AIRDROP_BALANCES_GZ="$SCRIPT_DIR/airdrop_balances.txt.gz"
AIRDROP_BALANCES_TXT="$WORK_DIR/airdrop_balances.txt"

if [ -f "$AIRDROP_BALANCES_GZ" ]; then
  printf "  Using cached airdrop balances\n"
else
  printf "  Downloading airdrop balances...\n"
  run curl -fsSL "$BALANCES_GZ_URL" -o "$AIRDROP_BALANCES_GZ"
fi
gzip -dkc "$AIRDROP_BALANCES_GZ" >"$AIRDROP_BALANCES_TXT"

# Merge airdrop with deployer balances by summing collisions. If an address
# appears in both the deployer balance sheet (fees consumed in genesis-mode
# txs) and the airdrop snapshot, `gnogenesis balances add` would otherwise
# replace the deployer entry with the airdrop entry — losing the deployer's
# residual / overwriting it. Summing preserves both contributions.
AIRDROP_MERGED_TXT="$WORK_DIR/airdrop_merged.txt"
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
collision_count=$(awk -F= 'FNR==NR{d[$1]=1;next} $1 in d{c++} END{print c+0}' \
  "$WORK_DIR_DEPLOYER_BALANCES" "$AIRDROP_BALANCES_TXT")
[ "$collision_count" -gt 0 ] && printf "  Merged %s deployer/airdrop collision(s)\n" "$collision_count"

airdrop_count=$(wc -l <"$AIRDROP_MERGED_TXT" | tr -d ' ')
printf "  Adding %s airdrop balances to genesis...\n" "$airdrop_count"
run "$GNOGENESIS_BIN" balances add -balance-sheet "$AIRDROP_MERGED_TXT" --genesis-path "$WORK_DIR_GENESIS" >/dev/null

# ---- 7. Append the 10 test-13 faucet balances at the tail of state.Balances
# `gnogenesis balances add` sorts state.Balances by Address.Compare. If we
# had added faucets in step 6, they would land at sort positions
# interspersed with the airdrop, shifting every airdrop entry that sorts
# after them by +1 per shift. That breaks the historical-tx SignerInfo
# invariant: validateSignerInfo (gno.land/pkg/gnoland/app.go:766)
# requires that state.Balances[i].Address matches whatever the cached
# txs.jsonl SignerInfo claims for accNum i — and those numbers come
# from gnoland1's launch ordering (e.g. manfred at 3,096,261).
#
# Note: the 7 NON-manfred deployer addresses ARE in step 6 — they were
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
# splice, so step 8 (verify) must not modify balances.

printf "\n=== Step 7/8: Appending 10 faucet balances to state.Balances tail ===\n"

# 7.1 Build the faucet balance sheet inline.
FAUCET_BALANCES_FILE="$WORK_DIR/faucet_balances.txt"
: >"$FAUCET_BALANCES_FILE"
for addr in "${FAUCET_ADDRESSES[@]}"; do
  echo "${addr}=${FAUCET_BALANCE}ugnot" >>"$FAUCET_BALANCES_FILE"
done
printf "  Faucet balance sheet built (%s entries, %s ugnot each)\n" \
  "${#FAUCET_ADDRESSES[@]}" "$FAUCET_BALANCE"

# 7.2 Throwaway genesis with ONLY the 10 faucets. The amino indenter is
# the same one the real genesis uses, so the resulting balance lines are
# format-compatible.
FAUCET_TMP_GENESIS="$WORK_DIR/faucet_tmp_genesis.json"
run "$GNOGENESIS_BIN" generate -chain-id "$CHAIN_ID" -genesis-time "$GENESIS_TIME" --output-path "$FAUCET_TMP_GENESIS" 2>&1 | sed 's/^/    /'
run "$GNOGENESIS_BIN" balances add -balance-sheet "$FAUCET_BALANCES_FILE" --genesis-path "$FAUCET_TMP_GENESIS" >/dev/null

# 7.3 Extract the 10 balance lines from the throwaway genesis. The
# state.Balances array is delimited by `    "balances": [` and `    ],`
# (4-space indent — `balances` is one level below the genesis root). The
# 10 lines come out preserving amino's trailing-comma rule: 9 entries
# with a comma, the last without.
FAUCET_EXTRAS_FILE="$WORK_DIR/faucet_extras.lines"
sed -n '/^    "balances": \[$/,/^    \],*$/{
    /^    "balances": \[$/d
    /^    \],*$/d
    p
}' "$FAUCET_TMP_GENESIS" >"$FAUCET_EXTRAS_FILE"
extras_count=$(wc -l <"$FAUCET_EXTRAS_FILE" | tr -d ' ')
if [ "$extras_count" -ne "${#FAUCET_ADDRESSES[@]}" ]; then
  echo "ERROR: extracted $extras_count balance lines from throwaway genesis, expected ${#FAUCET_ADDRESSES[@]}"
  exit 1
fi
printf "  Extracted %s amino-formatted balance lines from throwaway genesis\n" "$extras_count"

# 7.4 Splice the 10 lines into the real genesis. Awk passes the file
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
printf "  Splicing into state.Balances tail...\n"
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
printf "  Splice complete\n"

# ---- 8. Verify the generated genesis file

printf "\n=== Step 8/8: Verifying genesis ===\n"

run "$GNOGENESIS_BIN" verify -genesis-path "$WORK_DIR_GENESIS"
printf "  Verification passed\n"

cp "$WORK_DIR_GENESIS" "$GENESIS_FILE"
cp "$WORK_DIR/packages.gen.txt" "$SCRIPT_DIR/packages.gen.txt"

printf "\n=== Done ===\n"
printf "  sha256: %s\n" "$(shasum -a 256 "$GENESIS_FILE" | awk '{print $1}')"
printf "  -> out/base-genesis.json (%s)\n" "$(du -h "$GENESIS_FILE" | cut -f1)"
printf "  -> packages.gen.txt (tracked)\n"
printf "\n  Next: ./apply-test-13-replay.sh\n"
