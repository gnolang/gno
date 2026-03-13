#!/usr/bin/env bash
# Generate gnoland1 genesis.json.
#
# Usage:
#   ./gen-genesis.sh              # full build + genesis generation
#   ./gen-genesis.sh --debug      # show every command being run
#   ./gen-genesis.sh --txs-only   # stop after generating txs (skip balance calculation)
#   ./gen-genesis.sh --no-install # reuse previously built binaries
set -eo pipefail

# =============================================================================
# REVIEW THIS SECTION — update before each genesis generation.
# =============================================================================

# Packages to include in genesis (resolved with transitive dependencies).
# Use "..." suffix to match all sub-packages.
FILTERED_PACKAGES=(
  ./gno.land/r/sys/...
  ./gno.land/r/gov/...
  ./gno.land/r/gnoland/blog/...
  ./gno.land/r/gnoland/wugnot/...
  ./gno.land/r/gnoland/coins/...
  ./gno.land/r/gnoland/boards2/...
  ./gno.land/r/gnops/valopers/...
)

# Initial validator set. Format: "name power address pub_key"
# More validators can be added post-genesis via govDAO proposals (see govdao-scripts/add-validator.sh).
# 6 validators — BFT 2/3 threshold means 4 nodes must be up for consensus.
INITIAL_VALSET=(
  "gnocore-val-01 1 g1vta7dwp4guuhkfzksenfcheky4xf9hue8mgne4 gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpu5muc9ksphk3cayrduhathd2rw4talmtedpef3a44c2qfzzqalgl4c55y"
  "gnocore-val-02 1 g1d5hh9fw3l00gugfzafskaxqlmsyvxfaj6l2q60 gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpnj5vt2vkv94exe6cmdgqgtxmyfkvlhztnl0kj4xv97uz2t0muwe9mka0q"
  "moul-val-01 1 g1uhv7wr7nku89se3t7v8fpquc7n5sf8rfkywxpc gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqavtgten8l8k4f72j8klpu4l7tk2qw4kl8394krsaysmz2q0765ynvjag0"
  "aeddi-val-01 1 g10jdd8vlgydfypynrk23ul90jnsg5twrtvmcmh4 gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqve8jffvhy97sfc5gyvag09h8g9g3d9e4cta7s7m6vcmzug84kjywg7fn2y"
  "berty-val-01 1 g1jyaxj5t95dhlp9f8edkm0p0evw87qejluld86p gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq0wm4ysder0sgvre9qrahcz9fsg5qkdxuxetm5kmwaaul4e4e5p0rsx5f3"
  "samourai-val-01 1 g1kn7p0wqumvqlcqzhkwnavkhf0z4qnr73ltwsae gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpufsm93d5fmzxrug76esaxsdsmw0guy9e6geypw7ekz92sl3mte492q62e"
)

# Chain parameters.
CHAIN_ID=gnoland1
GENESIS_TIME=1770883200 # Thursday, February 12th 2026 09:00 GMT+0100 (CET)

# Airdrop balances (independence-day snapshot).
BALANCES_GZ_URL="https://github.com/gnolang/independence-day/raw/9dec38a4a72c9e84db7e78ae010370de250f2d64/mkgenesis/balances.txt.gz"

# =============================================================================
# INTERNAL — everything below is glue, you shouldn't need to change it.
# =============================================================================

# Deployer key mnemonic (deterministic — used only for genesis tx signing).
DEPLOYER_MNEMONIC="anchor hurt name seed oak spread anchor filter lesson shaft wasp home improve text behind toe segment lamp turn marriage female royal twice wealth"

# ---- Flags

STOP_AFTER_TXS_EXPORT=false
DEBUG=false
NO_INSTALL=false
GENESIS_FILE=genesis.json # set to absolute path below, after SCRIPT_DIR
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

# run executes a command, printing it first when --debug is set.
run() {
  if [ "$DEBUG" = true ]; then
    printf "    \033[2m\$ %s\033[0m\n" "$*" >&2
  fi
  "$@"
}

# Clean up background node on exit.
NODE_PID=""
cleanup() { [ -n "$NODE_PID" ] && kill "$NODE_PID" 2>/dev/null || true; }
trap cleanup EXIT

# ---- Derived paths (do not edit)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENESIS_FILE="$SCRIPT_DIR/$GENESIS_FILE"
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

# Clean up previous work directory (preserve bin/ when --no-install).
if [ "$NO_INSTALL" = true ]; then
  # Keep binaries, remove everything else.
  find "$WORK_DIR" -mindepth 1 -maxdepth 1 ! -name bin -exec rm -rf {} + 2>/dev/null || true
else
  rm -rf "$WORK_DIR"
fi

# ---- 1. Build binaries from source to ensure we have the right versions.

if [ "$NO_INSTALL" = true ]; then
  printf "\n=== Step 1/7: Skipping build (--no-install) ===\n"
  for bin in "$GNO_BIN" "$GNOKEY_BIN" "$GNOLAND_BIN" "$GNOGENESIS_BIN"; do
    if [ ! -x "$bin" ]; then
      echo "ERROR: --no-install but $bin not found. Run without --no-install first."
      exit 1
    fi
  done
else
  printf "\n=== Step 1/7: Building binaries ===\n"
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

printf "\n=== Step 2/7: Generating addpkg txs ===\n"

printf "  Resolving dependencies...\n"
pkg_dirs=$(cd "$EXAMPLES_DIR" && "$GNO_BIN" tool deplist -test-dep "${FILTERED_PACKAGES[@]}")
pkg_count=$(echo "$pkg_dirs" | wc -l | tr -d ' ')
printf "  Resolved %s packages in topological order\n" "$pkg_count"

# Save resolved package list for inspection.
{
  echo "# Generated by gen-genesis.sh — do not edit."
  echo "$pkg_dirs" | sed "s|$EXAMPLES_DIR/||g"
} >"$WORK_DIR/packages.gen.txt"

# Copy resolved packages into the working directory.
printf "  Copying packages to staging...\n"
WORK_DIR_EXAMPLES="$WORK_DIR/examples"
mkdir -p "$WORK_DIR_EXAMPLES"
while IFS= read -r dir; do
  [[ -z "$dir" ]] && continue
  rel="${dir#$EXAMPLES_DIR/}"
  mkdir -p "$(dirname "$WORK_DIR_EXAMPLES/$rel")"
  cp -r "$dir" "$WORK_DIR_EXAMPLES/$rel"
done <<<"$pkg_dirs"

# BUG: we need to figure out a way of not having to do this.
# Strip test files from staging — gnogenesis resolves all imports including
# from test files, which can pull in packages not in our set (e.g. r/tests/vm).
# The -test-dep flag above already ensured test *dependencies* (like uassert)
# are included; we just can't ship the test files themselves.
printf "  Stripping test files from staging...\n"
find "$WORK_DIR_EXAMPLES" -name '*_test.gno' -delete
find "$WORK_DIR_EXAMPLES" -name '*_filetest.gno' -delete

# Create deployer key (needed to sign MsgAddPackage and MsgRun txs).
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

# ---- 3. Generate setup transaction (validators, members, deployer cleanup)

printf "\n=== Step 3/7: Generating MsgRun setup tx (govdao_prop1.gno) ===\n"

SETUP_FILE="$SCRIPT_DIR/govdao_prop1.gno"

printf "  Generating MsgRun tx from %s...\n" "$(basename "$SETUP_FILE")"
SETUP_TX="$WORK_DIR/genesis_setup_tx.json"
SETUP_TX_FILE="$WORK_DIR/genesis_setup_tx.jsonl"
run "$GNOKEY_BIN" maketx run \
  --gas-wanted 100000000 \
  --gas-fee 1ugnot \
  --chainid "$CHAIN_ID" \
  --home "$WORK_DIR_GNOKEY_HOME" \
  GenesisDeployer \
  "$SETUP_FILE" >"$SETUP_TX"

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
  printf "  -> packages.gen.txt\n"
  printf "  -> genesis_txs.jsonl\n"
  exit 0
fi

# ---- 4. Calculate the deployers balances

printf "\n=== Step 4/7: Calculating deployer balances ===\n"

WORK_DIR_DEPLOYER_BALANCES="$WORK_DIR/deployers_balances.txt"
BALANCES_TMP_DIR="$WORK_DIR/balances-work"
BALANCES_TMP_FILE="$BALANCES_TMP_DIR/balances.txt"
BALANCES_TMP_GNOLAND_DATA="$BALANCES_TMP_DIR/gnoland-data"
BALANCES_TMP_GNOLAND_LOG="$BALANCES_TMP_DIR/node.log"
BALANCES_TMP_GENESIS="$BALANCES_TMP_DIR/genesis.json"
BALANCES_TMP_CREATOR_ADDRESSES="$BALANCES_TMP_DIR/gen-creators.txt"
INITIAL_BALANCE=1000000000000000
NODE_TIMEOUT=120
# Pick a free port for the temporary node (avoid collisions with running nodes).
NODE_RPC_PORT=$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()')
NODE_P2P_PORT=$((NODE_RPC_PORT + 1))
NODE_RPC_ADDR="127.0.0.1:$NODE_RPC_PORT"

rm -rf "$BALANCES_TMP_DIR"
mkdir -p "$BALANCES_TMP_DIR"

printf "  Extracting creator addresses...\n"
# Extract addresses from both MsgAddPackage ("creator") and MsgRun ("caller") txs.
grep -oE '"(creator|caller)":"[^"]*"' "$WORK_DIR_GENESIS_TXS" |
  sed 's/"creator":"//;s/"caller":"//;s/"//g' |
  sort -u >"$BALANCES_TMP_CREATOR_ADDRESSES"
addr_count=$(wc -l <"$BALANCES_TMP_CREATOR_ADDRESSES" | tr -d ' ')
printf "  Found %s unique creator/caller addresses\n" "$addr_count"

printf "  Generating over-provisioned balances...\n"
while IFS= read -r addr; do
  echo "${addr}=${INITIAL_BALANCE}ugnot" >>"$BALANCES_TMP_FILE"
done <"$BALANCES_TMP_CREATOR_ADDRESSES"

printf "  Setting up temporary node...\n"
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

printf "  Starting node (run 1: measure gas costs)...\n"
"$GNOLAND_BIN" start --skip-genesis-sig-verification -data-dir "$BALANCES_TMP_GNOLAND_DATA" -genesis "$BALANCES_TMP_GENESIS" >"$BALANCES_TMP_GNOLAND_LOG" 2>&1 &
NODE_PID=$!

elapsed=0
node_started=false
while [ "$elapsed" -lt "$NODE_TIMEOUT" ]; do
  if ! kill -0 "$NODE_PID" 2>/dev/null; then
    echo "ERROR: Node stopped unexpectedly. Last log lines:"
    tail -20 "$BALANCES_TMP_GNOLAND_LOG"
    exit 1
  fi
  if curl -sf "http://$NODE_RPC_ADDR/status" >/dev/null 2>&1; then
    node_started=true
    break
  fi
  sleep 1
  elapsed=$((elapsed + 1))
done

if [ "$node_started" = false ]; then
  kill "$NODE_PID" 2>/dev/null || true
  echo "ERROR: Node did not start within ${NODE_TIMEOUT}s. Last log lines:"
  tail -20 "$BALANCES_TMP_GNOLAND_LOG"
  exit 1
fi
printf "  Node ready (%ss)\n" "$elapsed"

printf "  Querying remaining balances...\n"
rm -f "$BALANCES_TMP_FILE"
while IFS= read -r addr; do
  remaining=""
  retry=0
  while [ "$retry" -lt "$NODE_TIMEOUT" ]; do
    if ! kill -0 "$NODE_PID" 2>/dev/null; then
      echo "ERROR: Node stopped unexpectedly during balance query. Last log lines:"
      tail -20 "$BALANCES_TMP_GNOLAND_LOG"
      exit 1
    fi
    query_output=$("$GNOKEY_BIN" query -remote "$NODE_RPC_ADDR" "bank/balances/$addr" 2>&1 || true)
    if echo "$query_output" | grep -q '^data:'; then
      remaining=$(echo "$query_output" | sed -n 's/.*"\([0-9]*\)ugnot".*/\1/p' | head -1)
      # Empty data field means 0 balance
      remaining=${remaining:-0}
      break
    fi
    sleep 1
    retry=$((retry + 1))
  done
  if [ -z "$remaining" ]; then
    echo "ERROR: Could not query balance for $addr after ${NODE_TIMEOUT}s. Last log lines:"
    tail -20 "$BALANCES_TMP_GNOLAND_LOG"
    kill "$NODE_PID" 2>/dev/null || true
    exit 1
  fi
  final=$((INITIAL_BALANCE - remaining))
  printf "    %s = %s ugnot\n" "$addr" "$final"
  echo "${addr}=${final}ugnot" >>"$BALANCES_TMP_FILE"
done <"$BALANCES_TMP_CREATOR_ADDRESSES"

kill "$NODE_PID" 2>/dev/null || true
wait "$NODE_PID" 2>/dev/null || true
NODE_PID=""

printf "  Setting up temporary node (run 2: verify)...\n"
rm -rf "$BALANCES_TMP_GNOLAND_DATA" "$BALANCES_TMP_GENESIS"

# Pick fresh ports for run 2 (previous ports may be in TIME_WAIT).
NODE_RPC_PORT=$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()')
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

printf "  Starting node (run 2: verify zero balances)...\n"
"$GNOLAND_BIN" start --skip-genesis-sig-verification -data-dir "$BALANCES_TMP_GNOLAND_DATA" -genesis "$BALANCES_TMP_GENESIS" >"$BALANCES_TMP_GNOLAND_LOG" 2>&1 &
NODE_PID=$!

elapsed=0
node_started=false
while [ "$elapsed" -lt "$NODE_TIMEOUT" ]; do
  if ! kill -0 "$NODE_PID" 2>/dev/null; then
    echo "ERROR: Node stopped unexpectedly. Last log lines:"
    tail -20 "$BALANCES_TMP_GNOLAND_LOG"
    exit 1
  fi
  if curl -sf "http://$NODE_RPC_ADDR/status" >/dev/null 2>&1; then
    node_started=true
    break
  fi
  sleep 1
  elapsed=$((elapsed + 1))
done

if [ "$node_started" = false ]; then
  kill "$NODE_PID" 2>/dev/null || true
  echo "ERROR: Node did not start within ${NODE_TIMEOUT}s. Last log lines:"
  tail -20 "$BALANCES_TMP_GNOLAND_LOG"
  exit 1
fi
printf "  Node ready (%ss)\n" "$elapsed"

printf "  Verifying all balances are zero...\n"
all_zero=true
while IFS= read -r addr; do
  remaining=""
  retry=0
  while [ "$retry" -lt "$NODE_TIMEOUT" ]; do
    if ! kill -0 "$NODE_PID" 2>/dev/null; then
      echo "ERROR: Node stopped unexpectedly during balance verification. Last log lines:"
      tail -20 "$BALANCES_TMP_GNOLAND_LOG"
      exit 1
    fi
    query_output=$("$GNOKEY_BIN" query -remote "$NODE_RPC_ADDR" "bank/balances/$addr" 2>&1 || true)
    if echo "$query_output" | grep -q '^data:'; then
      remaining=$(echo "$query_output" | sed -n 's/.*"\([0-9]*\)ugnot".*/\1/p' | head -1)
      # Empty data field means 0 balance
      remaining=${remaining:-0}
      break
    fi
    sleep 1
    retry=$((retry + 1))
  done
  if [ -z "$remaining" ]; then
    echo "ERROR: Could not query balance for $addr after ${NODE_TIMEOUT}s. Last log lines:"
    tail -20 "$BALANCES_TMP_GNOLAND_LOG"
    kill "$NODE_PID" 2>/dev/null || true
    exit 1
  fi
  if [ "$remaining" -ne 0 ]; then
    printf "    FAIL: %s has %sugnot remaining\n" "$addr" "$remaining"
    all_zero=false
  else
    printf "    ok: %s\n" "$addr"
  fi
done <"$BALANCES_TMP_CREATOR_ADDRESSES"

kill "$NODE_PID" 2>/dev/null || true
wait "$NODE_PID" 2>/dev/null || true
NODE_PID=""

if [ "$all_zero" = true ]; then
  printf "  All balances zero — deployer costs verified\n"
  cp "$BALANCES_TMP_FILE" "$WORK_DIR_DEPLOYER_BALANCES"
else
  echo "ERROR: Some balances are not zero after replay. Check $BALANCES_TMP_FILE."
  exit 1
fi

printf "  Adding deployer balances to genesis...\n"
run "$GNOGENESIS_BIN" balances add -balance-sheet "$WORK_DIR_DEPLOYER_BALANCES" --genesis-path "$WORK_DIR_GENESIS" >/dev/null

# ---- 5. Download and add the airdrop balances

printf "\n=== Step 5/7: Downloading airdrop balances ===\n"

AIRDROP_BALANCES_GZ="$WORK_DIR/airdrop_balances.txt.gz"
AIRDROP_BALANCES_TXT="$WORK_DIR/airdrop_balances.txt"

printf "  Downloading...\n"
run curl -fsSL "$BALANCES_GZ_URL" -o "$AIRDROP_BALANCES_GZ"
gzip -dc "$AIRDROP_BALANCES_GZ" >"$AIRDROP_BALANCES_TXT"

airdrop_count=$(wc -l <"$AIRDROP_BALANCES_TXT" | tr -d ' ')
# TODO: We need to verify if there is a colision between deployer and airdrop addresses.
# See: https://github.com/gnolang/gno/pull/5250/changes#discussion_r2925485031
printf "  Adding %s airdrop balances to genesis...\n" "$airdrop_count"
run "$GNOGENESIS_BIN" balances add -balance-sheet "$AIRDROP_BALANCES_TXT" --genesis-path "$WORK_DIR_GENESIS" >/dev/null

# ---- 6. Add the initial validator set to the genesis file

printf "\n=== Step 6/7: Adding validators ===\n"

for validator in "${INITIAL_VALSET[@]}"; do
  read -r name power address pub_key <<<"$validator"
  printf "  %s (power=%s, %s)\n" "$name" "$power" "$address"
  run "$GNOGENESIS_BIN" validator add -name "$name" -power "$power" -address "$address" -pub-key "$pub_key" --genesis-path "$WORK_DIR_GENESIS"
done

# ---- 7. Verify the generated genesis file

printf "\n=== Step 7/7: Verifying genesis ===\n"

run "$GNOGENESIS_BIN" verify -genesis-path "$WORK_DIR_GENESIS"
printf "  Verification passed\n"

cp "$WORK_DIR_GENESIS" "$GENESIS_FILE"
cp "$WORK_DIR/packages.gen.txt" "$SCRIPT_DIR/packages.gen.txt"
cp "$WORK_DIR_GENESIS_TXS" "$SCRIPT_DIR/genesis_txs.jsonl"

# Generate a redacted genesis (no airdrop balances) for version control.
jq '.app_state.balances = (.app_state.balances[:10])' "$GENESIS_FILE" >"$SCRIPT_DIR/genesis-redacted.json"

printf "\n=== Done ===\n"
printf "  sha256: %s\n" "$(shasum -a 256 "$GENESIS_FILE" | awk '{print $1}')"
printf "  -> genesis.json (gitignored, %s)\n" "$(du -h "$GENESIS_FILE" | cut -f1)"
printf "  -> genesis-redacted.json (%s)\n" "$(du -h "$SCRIPT_DIR/genesis-redacted.json" | cut -f1)"
printf "  -> packages.gen.txt\n"
printf "  -> genesis_txs.jsonl\n"
