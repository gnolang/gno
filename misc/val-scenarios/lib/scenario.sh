#!/usr/bin/env bash
set -euo pipefail

if [ "${BASH_VERSINFO[0]}" -lt 4 ]; then
  printf 'error: bash 4+ required (found %s); install with: brew install bash\n' "$BASH_VERSION" >&2
  exit 1
fi

SCENARIO_SELF="${BASH_SOURCE[0]}"
SCENARIO_LIB_DIR="$(cd "$(dirname "${SCENARIO_SELF}")" && pwd)"
REPO_ROOT="$(cd "${SCENARIO_LIB_DIR}/../../.." && pwd)"

# RUNTIME selects how nodes and one-shot CLI commands are executed:
#   docker (default) — every binary runs in a container; nodes address each
#     other by docker-compose service name on fixed ports 26656/26657.
#   local            — the four binaries (gnoland, gnokey, gnogenesis,
#     valsignerd) run directly from BIN_DIR; nodes bind 127.0.0.1 on
#     deterministic per-node ports and are supervised as background processes.
# Build the local binaries with `make build-binaries`.
RUNTIME="${RUNTIME:-docker}"
BIN_DIR="${BIN_DIR:-${SCENARIO_LIB_DIR}/../bin}"
# Local-runtime binaries, produced by `make build-binaries` into BIN_DIR.
GNOLAND_BIN="${BIN_DIR}/gnoland"
GNOKEY_BIN="${BIN_DIR}/gnokey"
GNOGENESIS_BIN="${BIN_DIR}/gnogenesis"
VALSIGNERD_BIN="${BIN_DIR}/valsignerd"
# Base host ports for the local runtime. Each node n (0-indexed) binds
# RPC=base+n, P2P=p2p_base+n, remote-signer=rs_base+n, control=ctrl_base+n.
# Override the bases to run multiple local scenarios concurrently.
LOCAL_RPC_PORT_BASE="${LOCAL_RPC_PORT_BASE:-26700}"
LOCAL_P2P_PORT_BASE="${LOCAL_P2P_PORT_BASE:-26800}"
LOCAL_RS_PORT_BASE="${LOCAL_RS_PORT_BASE:-26900}"
LOCAL_CONTROL_PORT_BASE="${LOCAL_CONTROL_PORT_BASE:-28080}"

IMAGE_NAME="${IMAGE_NAME:-gno-val-scenario-core:local}"
GNOKEY_IMAGE="${GNOKEY_IMAGE:-${IMAGE_NAME}}"
GNOGENESIS_IMAGE="${GNOGENESIS_IMAGE:-gnogenesis:local}"
VALSIGNER_IMAGE="${VALSIGNER_IMAGE:-valsignerd:local}"
GNO_ROOT="${GNO_ROOT:-${REPO_ROOT}}"
WORK_ROOT="${WORK_ROOT:-/tmp/gno-val-tests}"
CHAIN_ID="${CHAIN_ID:-dev}"
TIMEOUT_COMMIT="${TIMEOUT_COMMIT:-1s}"
LOG_LEVEL="${LOG_LEVEL:-info}"
REMOTE_SIGNER_REQUEST_TIMEOUT="${REMOTE_SIGNER_REQUEST_TIMEOUT:-30s}"
TX_KEY_NAME="${TX_KEY_NAME:-scenario-tx}"
TX_PASSWORD="${TX_PASSWORD:-test123456}"
TX_MNEMONIC="${TX_MNEMONIC:-source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast}"
TX_ADDRESS="${TX_ADDRESS:-g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5}"
TX_BALANCE="${TX_BALANCE:-100000000000ugnot}"
TX_GAS_FEE="${TX_GAS_FEE:-1000000ugnot}"
TX_GAS_WANTED_ADD_PKG="${TX_GAS_WANTED_ADD_PKG:-50000000}"
TX_GAS_WANTED_CALL="${TX_GAS_WANTED_CALL:-3000000}"
TX_GAS_WANTED_RUN="${TX_GAS_WANTED_RUN:-5000000}"
TX_GAS_WANTED_SEND="${TX_GAS_WANTED_SEND:-2000000}"

declare -a SCENARIO_NODES=()
declare -a SCENARIO_VALIDATORS=()
declare -a SCENARIO_GENESIS_VALIDATORS=()
declare -a SCENARIO_SENTRIES=()
declare -a SCENARIO_SIGNERS=()
declare -A NODE_ROLE=()
declare -A NODE_SERVICE=()
declare -A NODE_MONIKER=()
declare -A NODE_RPC_PORT=()
declare -A NODE_PEX=()
declare -A NODE_SENTRY=()
declare -A NODE_ID=()
declare -A NODE_ADDRESS=()
declare -A NODE_PUBKEY=()
declare -A NODE_DATA_DIR=()
declare -A NODE_POWER=()
declare -A NODE_CONTROLLABLE_SIGNER=()
declare -A NODE_SIGNER_SERVICE=()
declare -A NODE_CONTROL_PORT=()
declare -A NODE_LOG_PID=()
# Local-runtime state. Ports start at base+index and shift up to the next free
# port if it is taken (see _pick_free_port), keyed off NODE_COUNTER at register.
declare -A NODE_P2P_PORT=()   # local host P2P port
declare -A NODE_RS_PORT=()    # local host remote-signer port (controllable signers)
declare -A NODE_PID=()        # local gnoland process pid
declare -A NODE_SIGNER_PID=() # local valsignerd process pid
declare -A _CLAIMED_PORTS=()  # ports already handed out this scenario (local)
_PICKED_PORT=""               # out-param for _pick_free_port
NODE_COUNTER=0

SCENARIO_NAME=""
PROJECT_NAME=""
SCENARIO_DIR=""
COMPOSE_FILE=""
KEY_HOME=""
NETWORK_NAME=""

log() {
  printf '[%s] %s\n' "${SCENARIO_NAME:-scenario}" "$*"
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

join_by() {
  local delimiter="${1:?delimiter required}"
  shift || true
  local out=""
  local first=1
  local value
  for value in "$@"; do
    if [ "$first" -eq 1 ]; then
      out="$value"
      first=0
    else
      out="${out}${delimiter}${value}"
    fi
  done
  printf '%s' "$out"
}

slugify() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9' '-'
}

# _pick_free_port START — first free 127.0.0.1 TCP port at or above START that
# this scenario has not already handed out. Used at register time (local
# runtime) so the deterministic base+index port shifts up when something else
# holds it, instead of failing later with a 120s wait_for_rpc timeout. The port
# must be chosen before prepare_network bakes it into genesis/peers/inventory,
# so this runs at registration, not at start. Probes via bash's built-in
# /dev/tcp (no nc/lsof dependency): a successful connect means someone is
# listening, so the port is busy.
#
# The result is returned in the global _PICKED_PORT rather than on stdout,
# because callers must NOT use $(...) here: command substitution runs in a
# subshell, which would discard the _CLAIMED_PORTS update and let sibling ports
# collide.
_pick_free_port() {
  local p="${1:?start port required}"
  # The connect probe runs in a subshell, so fd 3 is scoped to it and needs no
  # explicit close here.
  while [ -n "${_CLAIMED_PORTS[$p]:-}" ] || (exec 3<>"/dev/tcp/127.0.0.1/$p") 2>/dev/null; do
    p=$((p + 1))
  done
  _CLAIMED_PORTS[$p]=1
  _PICKED_PORT="$p"
}

require_tools() {
  local missing=()
  local tool
  local -a tools=(jq curl)
  [ "$RUNTIME" = "docker" ] && tools+=(docker)
  for tool in "${tools[@]}"; do
    if ! command -v "$tool" >/dev/null 2>&1; then
      missing+=("$tool")
    fi
  done
  if [ "${#missing[@]}" -gt 0 ]; then
    die "missing required tools: $(join_by ', ' "${missing[@]}")"
  fi
}

scenario_init() {
  local name="${1:?scenario name required}"

  SCENARIO_NAME="$name"
  PROJECT_NAME="$(slugify "$name")"
  SCENARIO_DIR="${WORK_ROOT}/${PROJECT_NAME}"
  COMPOSE_FILE="${SCENARIO_DIR}/docker-compose.yml"
  KEY_HOME="${SCENARIO_DIR}/keys"
  NETWORK_NAME="${PROJECT_NAME}_chain"

  log "scenario dir: ${SCENARIO_DIR}"

  SCENARIO_NODES=()
  SCENARIO_VALIDATORS=()
  SCENARIO_GENESIS_VALIDATORS=()
  SCENARIO_SENTRIES=()
  SCENARIO_SIGNERS=()
  NODE_ROLE=()
  NODE_SERVICE=()
  NODE_MONIKER=()
  NODE_RPC_PORT=()
  NODE_PEX=()
  NODE_SENTRY=()
  NODE_ID=()
  NODE_ADDRESS=()
  NODE_PUBKEY=()
  NODE_DATA_DIR=()
  NODE_POWER=()
  NODE_CONTROLLABLE_SIGNER=()
  NODE_SIGNER_SERVICE=()
  NODE_CONTROL_PORT=()
  NODE_LOG_PID=()
  NODE_P2P_PORT=()
  NODE_RS_PORT=()
  NODE_PID=()
  NODE_SIGNER_PID=()
  _CLAIMED_PORTS=()
  NODE_COUNTER=0
}

register_node() {
  local name="${1:?node name required}"
  local role="${2:?role required}"
  local rpc_port="${3:-}"
  local pex="${4:?pex required}"
  local sentry="${5:-}"
  local in_genesis="${6:-true}"

  [ -z "${NODE_ROLE[$name]:-}" ] || die "node ${name} already exists"

  SCENARIO_NODES+=("$name")
  NODE_ROLE[$name]="$role"
  NODE_SERVICE[$name]="$name"
  NODE_MONIKER[$name]="$name"
  NODE_PEX[$name]="$pex"
  NODE_SENTRY[$name]="$sentry"
  if [ "$RUNTIME" = "local" ]; then
    # The docker --rpc-port hint is irrelevant locally; assign per-node host
    # ports starting from base+index so nodes can address each other on
    # 127.0.0.1, shifting up to the next free port when a candidate is taken
    # (e.g. a concurrent local scenario). RS/control are only used by
    # controllable signers. Chosen here, before prepare_network bakes them into
    # genesis/peers/inventory.
    # Not $(...) — _pick_free_port mutates _CLAIMED_PORTS in this shell and
    # returns via _PICKED_PORT (see its comment).
    _pick_free_port "$((LOCAL_RPC_PORT_BASE + NODE_COUNTER))";     NODE_RPC_PORT[$name]="$_PICKED_PORT"
    _pick_free_port "$((LOCAL_P2P_PORT_BASE + NODE_COUNTER))";     NODE_P2P_PORT[$name]="$_PICKED_PORT"
    _pick_free_port "$((LOCAL_RS_PORT_BASE + NODE_COUNTER))";      NODE_RS_PORT[$name]="$_PICKED_PORT"
    _pick_free_port "$((LOCAL_CONTROL_PORT_BASE + NODE_COUNTER))"; NODE_CONTROL_PORT[$name]="$_PICKED_PORT"
  else
    NODE_RPC_PORT[$name]="$rpc_port"
  fi
  NODE_COUNTER=$((NODE_COUNTER + 1))

  case "$role" in
    validator)
      SCENARIO_VALIDATORS+=("$name")
      if [ "$in_genesis" = "true" ]; then
        SCENARIO_GENESIS_VALIDATORS+=("$name")
      fi
      ;;
    sentry) SCENARIO_SENTRIES+=("$name") ;;
    *) die "unsupported node role ${role}" ;;
  esac
}

gen_validator() {
  local name="${1:?validator name required}"
  shift || true

  local rpc_port=""
  local sentry=""
  local pex="true"
  local power="1"
  local controllable_signer="false"
  local in_genesis="true"

  while [ "$#" -gt 0 ]; do
    case "$1" in
      --rpc-port)
        rpc_port="${2:?missing rpc port}"
        shift 2
        ;;
      --sentry)
        sentry="${2:?missing sentry name}"
        pex="false"
        shift 2
        ;;
      --pex)
        pex="${2:?missing pex value}"
        shift 2
        ;;
      --power)
        power="${2:?missing power value}"
        shift 2
        ;;
      --controllable-signer)
        controllable_signer="true"
        shift
        ;;
      --not-in-genesis)
        in_genesis="false"
        shift
        ;;
      *)
        die "unknown gen_validator option: $1"
        ;;
    esac
  done

  register_node "$name" validator "$rpc_port" "$pex" "$sentry" "$in_genesis"
  NODE_POWER[$name]="$power"
  NODE_CONTROLLABLE_SIGNER[$name]="$controllable_signer"
  if [ "$controllable_signer" = "true" ]; then
    NODE_SIGNER_SERVICE[$name]="${name}-signer"
    # Local control/RS ports are assigned in register_node; docker resolves the
    # control port from the ephemeral host mapping after start.
    [ "$RUNTIME" = "local" ] || NODE_CONTROL_PORT[$name]=""
    SCENARIO_SIGNERS+=("$name")
  fi
}

gen_sentry() {
  local name="${1:?sentry name required}"
  shift || true

  local rpc_port=""
  local pex="false"

  while [ "$#" -gt 0 ]; do
    case "$1" in
      --rpc-port)
        rpc_port="${2:?missing rpc port}"
        shift 2
        ;;
      --pex)
        pex="${2:?missing pex value}"
        shift 2
        ;;
      *)
        die "unknown gen_sentry option: $1"
        ;;
    esac
  done

  register_node "$name" sentry "$rpc_port" "$pex" ""
}

ensure_runtime_ready() {
  if [ "$RUNTIME" = "local" ]; then
    local -a needed=("$GNOLAND_BIN" "$GNOKEY_BIN" "$GNOGENESIS_BIN")
    [ "${#SCENARIO_SIGNERS[@]}" -gt 0 ] && needed+=("$VALSIGNERD_BIN")
    local bin
    for bin in "${needed[@]}"; do
      [ -x "$bin" ] || die "binary not found or not executable: ${bin}; run \`make build-binaries\` first"
    done
    return 0
  fi

  local image_id
  image_id="$(docker images -q "$IMAGE_NAME" 2>/dev/null)"
  if [ -z "$image_id" ]; then
    die "docker image ${IMAGE_NAME} not found; run \`make build-images\` first"
  fi
  image_id="$(docker images -q "$GNOKEY_IMAGE" 2>/dev/null)"
  if [ -z "$image_id" ]; then
    die "docker image ${GNOKEY_IMAGE} not found; run \`make build-images\` first"
  fi
  image_id="$(docker images -q "$GNOGENESIS_IMAGE" 2>/dev/null)"
  if [ -z "$image_id" ]; then
    die "docker image ${GNOGENESIS_IMAGE} not found; run \`make build-images\` first"
  fi
  if [ "${#SCENARIO_SIGNERS[@]}" -gt 0 ]; then
    image_id="$(docker images -q "$VALSIGNER_IMAGE" 2>/dev/null)"
    if [ -z "$image_id" ]; then
      die "docker image ${VALSIGNER_IMAGE} not found; run \`make build-images\` first"
    fi
  fi
}

compose() {
  docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" "$@"
}

# ---------------------------------------------------------------------------
# Runtime path accessors and command runners.
#
# In docker the binaries see container paths (/data, /work, /gnoroot, /keys)
# bound from host dirs; in local mode every "container path" is just the host
# path. The dpath_* helpers return the path a binary should be handed, and the
# run_* helpers execute the binary (in a container, or directly).
# ---------------------------------------------------------------------------

dpath_data() { # data dir for a node as seen by the binary
  if [ "$RUNTIME" = "local" ]; then printf '%s' "${NODE_DATA_DIR[$1]}"; else printf '/data'; fi
}
dpath_work() { # scenario work dir as seen by the binary
  if [ "$RUNTIME" = "local" ]; then printf '%s' "$SCENARIO_DIR"; else printf '/work'; fi
}
dpath_gnoroot() { # GNO_ROOT as seen by the binary
  if [ "$RUNTIME" = "local" ]; then printf '%s' "$GNO_ROOT"; else printf '/gnoroot'; fi
}
# dmount HOSTPATH CONTPATH — path a binary should use for a bound dir: the host
# path locally, the container path in docker.
dmount() {
  if [ "$RUNTIME" = "local" ]; then printf '%s' "$1"; else printf '%s' "$2"; fi
}
dpath_keys() { # a host keys home as seen by the binary (bound at /keys in docker)
  dmount "$1" /keys
}

# run_gnoland_node NODE ARGS... — one-shot gnoland for a node's data dir.
run_gnoland_node() {
  local node="${1:?node required}"; shift
  if [ "$RUNTIME" = "local" ]; then
    "$GNOLAND_BIN" "$@"
  else
    docker run --rm --entrypoint /usr/bin/gnoland \
      -v "${NODE_DATA_DIR[$node]}:/data" "$IMAGE_NAME" "$@"
  fi
}

# run_gnogenesis [-v HOST:CONT]... ARGS... — gnogenesis with work+gnoroot bound.
# Extra -v binds (docker only) may precede the subcommand; they are dropped
# locally since paths are already host paths.
run_gnogenesis() {
  local -a binds=()
  while [ "$#" -gt 0 ] && [ "$1" = "-v" ]; do
    binds+=("-v" "$2"); shift 2
  done
  if [ "$RUNTIME" = "local" ]; then
    "$GNOGENESIS_BIN" "$@"
  else
    docker run --rm --entrypoint /usr/bin/gnogenesis \
      -v "${SCENARIO_DIR}:/work" -v "${GNO_ROOT}:/gnoroot:ro" \
      "${binds[@]}" "$GNOGENESIS_IMAGE" "$@"
  fi
}

# run_gnokey [-v HOST:CONT]... ARGS... — gnokey, reading stdin. Work dir is
# bound in docker; extra -v binds (e.g. a keys home or pkg dir) may precede the
# subcommand.
run_gnokey() {
  local -a binds=()
  while [ "$#" -gt 0 ] && [ "$1" = "-v" ]; do
    binds+=("-v" "$2"); shift 2
  done
  if [ "$RUNTIME" = "local" ]; then
    "$GNOKEY_BIN" "$@"
  else
    docker run -i --rm --entrypoint /usr/bin/gnokey \
      -v "${SCENARIO_DIR}:/work" "${binds[@]}" "$GNOKEY_IMAGE" "$@"
  fi
}

init_node_dirs() {
  local node
  for node in "${SCENARIO_NODES[@]}"; do
    local node_dir="${SCENARIO_DIR}/nodes/${node}"
    NODE_DATA_DIR[$node]="$node_dir"
    mkdir -p "$node_dir"

    local data
    data="$(dpath_data "$node")"
    run_gnoland_node "$node" secrets init --data-dir "${data}/secrets" >/dev/null
    run_gnoland_node "$node" config init --config-path "${data}/config/config.toml" >/dev/null
  done
}

collect_node_ids() {
  local node secrets
  for node in "${SCENARIO_NODES[@]}"; do
    secrets="$(dpath_data "$node")/secrets"
    NODE_ID[$node]="$(run_gnoland_node "$node" secrets get node_id.id --data-dir "$secrets" --raw | tr -d '\r\n')"
    NODE_ADDRESS[$node]="$(run_gnoland_node "$node" secrets get validator_key.address --data-dir "$secrets" --raw | tr -d '\r\n')"
    NODE_PUBKEY[$node]="$(run_gnoland_node "$node" secrets get validator_key.pub_key --data-dir "$secrets" --raw | tr -d '\r\n')"
  done
}

generate_genesis() {
  [ "${#SCENARIO_GENESIS_VALIDATORS[@]}" -gt 0 ] || die "at least one genesis validator is required"
  [ -d "${GNO_ROOT}/examples" ] || die "GNO_ROOT examples not found at ${GNO_ROOT}/examples; run 'make clone-gno' or set GNO_ROOT"

  local genesis_work="${SCENARIO_DIR}/genesis-work"
  local gnokey_home="${genesis_work}/gnokey-home"
  local deployer_name="GenesisDeployer"
  # Same mnemonic as gen-genesis.sh; address = g1edq4dugw0sgat4zxcw9xardvuydqf6cgleuc8p
  local deployer_mnemonic="anchor hurt name seed oak spread anchor filter lesson shaft wasp home improve text behind toe segment lamp turn marriage female royal twice wealth"

  mkdir -p "$genesis_work" "$gnokey_home"

  # Paths as the binaries see them (host paths locally, container paths in docker).
  local work gnoroot keys genesis
  work="$(dpath_work)"
  gnoroot="$(dpath_gnoroot)"
  keys="$(dpath_keys "$gnokey_home")"
  genesis="${work}/genesis.json"

  log "creating genesis deployer key"
  printf '%s\n\n' "$deployer_mnemonic" | \
    run_gnokey -v "${gnokey_home}:/keys" \
      add --recover "$deployer_name" --home "$keys" --insecure-password-stdin >/dev/null

  log "generating empty genesis"
  run_gnogenesis generate \
    --chain-id "$CHAIN_ID" \
    --genesis-time "$(date +%s)" \
    --output-path "$genesis" >/dev/null

  log "adding packages from GNO_ROOT"
  printf '\n' | \
    run_gnogenesis -v "${gnokey_home}:/keys" \
      txs add packages "${gnoroot}/examples" \
        --genesis-path "$genesis" \
        --gno-home "$keys" \
        --key-name "$deployer_name" \
        --insecure-password-stdin >/dev/null

  log "generating valset-init MsgRun"
  local valset_file="${genesis_work}/valset-init.gno"
  local valset_entries=""
  local node
  for node in "${SCENARIO_GENESIS_VALIDATORS[@]}"; do
    valset_entries+="$(printf '\t\t\t\t{Address: address("%s"), PubKey: "%s", VotingPower: %s},\n' \
      "${NODE_ADDRESS[$node]}" "${NODE_PUBKEY[$node]}" "${NODE_POWER[$node]:-1}")"
  done
  awk -v entries="$valset_entries" \
    '/\/\/ GEN:VALSET/ { printf "%s", entries; next } { print }' \
    "${SCENARIO_LIB_DIR}/valset-init.gno.tpl" > "$valset_file"

  local setup_tx="${genesis_work}/valset-init-tx.json"
  local setup_tx_jsonl="${genesis_work}/valset-init-tx.jsonl"

  printf '\n' | run_gnokey -v "${gnokey_home}:/keys" \
    maketx run \
      --gas-wanted 100000000 \
      --gas-fee 1ugnot \
      --chainid "$CHAIN_ID" \
      --broadcast=false \
      --home "$keys" \
      --insecure-password-stdin \
      "$deployer_name" \
      "${work}/genesis-work/valset-init.gno" > "$setup_tx"

  printf '\n' | run_gnokey -v "${gnokey_home}:/keys" \
    sign \
      --tx-path "${work}/genesis-work/valset-init-tx.json" \
      --chainid "$CHAIN_ID" \
      --account-number 0 \
      --account-sequence 0 \
      --home "$keys" \
      --insecure-password-stdin \
      "$deployer_name" >/dev/null

  jq -c '{tx: .}' < "$setup_tx" > "$setup_tx_jsonl"

  run_gnogenesis txs add sheets --genesis-path "$genesis" "${work}/genesis-work/valset-init-tx.jsonl" >/dev/null

  log "adding ${#SCENARIO_GENESIS_VALIDATORS[@]} validators to consensus layer"
  for node in "${SCENARIO_GENESIS_VALIDATORS[@]}"; do
    run_gnogenesis validator add \
      --genesis-path "$genesis" \
      --name "$node" \
      --address "${NODE_ADDRESS[$node]}" \
      --pub-key "${NODE_PUBKEY[$node]}" \
      --power "${NODE_POWER[$node]:-1}" >/dev/null
  done

  log "adding test1 balance"
  run_gnogenesis balances add --genesis-path "$genesis" --single "${TX_ADDRESS}=${TX_BALANCE}" >/dev/null

  local genesis_file="${SCENARIO_DIR}/genesis.json"
  for node in "${SCENARIO_NODES[@]}"; do
    cp "$genesis_file" "${NODE_DATA_DIR[$node]}/genesis.json"
  done
}

format_peer_entry() {
  local node="${1:?node required}"
  if [ "$RUNTIME" = "local" ]; then
    printf '%s@127.0.0.1:%s' "${NODE_ID[$node]}" "${NODE_P2P_PORT[$node]}"
  else
    printf '%s@%s:26656' "${NODE_ID[$node]}" "${NODE_SERVICE[$node]}"
  fi
}

persistent_peer_targets() {
  local node="${1:?node required}"
  local role="${NODE_ROLE[$node]}"
  local target
  local -a peers=()

  case "$role" in
    validator)
      if [ -n "${NODE_SENTRY[$node]}" ]; then
        peers+=("${NODE_SENTRY[$node]}")
      else
        for target in "${SCENARIO_VALIDATORS[@]}"; do
          if [ "$target" != "$node" ] && [ -z "${NODE_SENTRY[$target]}" ]; then
            peers+=("$target")
          fi
        done
        for target in "${SCENARIO_SENTRIES[@]}"; do
          peers+=("$target")
        done
      fi
      ;;
    sentry)
      for target in "${SCENARIO_VALIDATORS[@]}"; do
        if [ "$target" = "$node" ]; then
          continue
        fi
        # Only peer with validators that are not hidden behind this sentry.
        # Hidden validators dial the sentry themselves and are listed in
        # private_peer_ids; they must not appear in persistent_peers/seeds.
        if [ -z "${NODE_SENTRY[$target]}" ]; then
          peers+=("$target")
        fi
      done
      for target in "${SCENARIO_SENTRIES[@]}"; do
        if [ "$target" != "$node" ]; then
          peers+=("$target")
        fi
      done
      ;;
    *)
      die "unsupported role ${role}"
      ;;
  esac

  printf '%s\n' "${peers[@]}" | awk '!seen[$0]++ && NF'
}

persistent_peers_for_node() {
  local node="${1:?node required}"
  local -a rendered=()
  local target

  while IFS= read -r target; do
    [ -n "$target" ] || continue
    rendered+=("$(format_peer_entry "$target")")
  done < <(persistent_peer_targets "$node")

  join_by ',' "${rendered[@]}"
}

set_config_value() {
  local node="${1:?node required}"
  local key="${2:?config key required}"
  local value="${3:?config value required}"

  run_gnoland_node "$node" \
    config set \
      --config-path "$(dpath_data "$node")/config/config.toml" \
      "$key" "$value" >/dev/null
}

private_peer_ids_for_sentry() {
  local sentry="${1:?sentry required}"
  local -a ids=()
  local target
  for target in "${SCENARIO_VALIDATORS[@]}"; do
    if [ "${NODE_SENTRY[$target]}" = "$sentry" ]; then
      ids+=("${NODE_ID[$target]}")
    fi
  done
  join_by ',' "${ids[@]}"
}

# apply_peer_config NODE — (re)write a node's persistent_peers/seeds from the
# current peer graph. Shared by initial configuration and local port rotation.
apply_peer_config() {
  local node="${1:?node required}"
  local peers
  peers="$(persistent_peers_for_node "$node")"
  if [ -n "$peers" ]; then
    set_config_value "$node" p2p.persistent_peers "$peers"
    set_config_value "$node" p2p.seeds "$peers"
  fi
}

configure_nodes() {
  local node
  for node in "${SCENARIO_NODES[@]}"; do
    # Docker binds 0.0.0.0 on fixed ports inside each container's own netns;
    # locally every node shares 127.0.0.1 so each needs a distinct port.
    local rpc_laddr p2p_laddr rs_addr
    if [ "$RUNTIME" = "local" ]; then
      rpc_laddr="tcp://127.0.0.1:${NODE_RPC_PORT[$node]}"
      p2p_laddr="tcp://127.0.0.1:${NODE_P2P_PORT[$node]}"
    else
      rpc_laddr="tcp://0.0.0.0:26657"
      p2p_laddr="tcp://0.0.0.0:26656"
    fi

    set_config_value "$node" moniker "${NODE_MONIKER[$node]}"
    set_config_value "$node" rpc.laddr "$rpc_laddr"
    set_config_value "$node" p2p.laddr "$p2p_laddr"
    set_config_value "$node" p2p.pex "${NODE_PEX[$node]}"
    apply_peer_config "$node"
    set_config_value "$node" consensus.timeout_commit "$TIMEOUT_COMMIT"
    if [ "${NODE_CONTROLLABLE_SIGNER[$node]:-false}" = "true" ]; then
      if [ "$RUNTIME" = "local" ]; then
        rs_addr="tcp://127.0.0.1:${NODE_RS_PORT[$node]}"
      else
        rs_addr="tcp://${NODE_SIGNER_SERVICE[$node]}:26659"
      fi
      set_config_value "$node" consensus.priv_validator.remote_signer.server_address "$rs_addr"
      set_config_value "$node" consensus.priv_validator.remote_signer.request_timeout "$REMOTE_SIGNER_REQUEST_TIMEOUT"
    fi

    if [ "${NODE_ROLE[$node]}" = "sentry" ]; then
      local private_ids
      private_ids="$(private_peer_ids_for_sentry "$node")"
      if [ -n "$private_ids" ]; then
        set_config_value "$node" p2p.private_peer_ids "$private_ids"
      fi
    fi
  done
}

write_compose_file() {
  {
    printf 'name: %s\n\n' "$PROJECT_NAME"
    printf 'services:\n'
    local node
    local signer
    for signer in "${SCENARIO_SIGNERS[@]}"; do
      printf '  %s:\n' "${NODE_SIGNER_SERVICE[$signer]}"
      printf '    image: "%s"\n' "$VALSIGNER_IMAGE"
      printf '    command:\n'
      printf '      - --key-file\n'
      printf '      - /data/secrets/priv_validator_key.json\n'
      printf '      - --listen-addr\n'
      printf '      - :8080\n'
      printf '      - --remote-signer-addr\n'
      printf '      - tcp://0.0.0.0:26659\n'
      printf '    volumes:\n'
      printf '      - "%s:/data:ro"\n' "${NODE_DATA_DIR[$signer]}"
      printf '    ports:\n'
      if [ -n "${NODE_CONTROL_PORT[$signer]:-}" ]; then
        printf '      - "%s:8080"\n' "${NODE_CONTROL_PORT[$signer]}"
      else
        printf '      - "::8080"\n'
      fi
      printf '    networks:\n'
      printf '      - chain\n'
      printf '    stop_grace_period: 5s\n'
    done

    for node in "${SCENARIO_NODES[@]}"; do
      printf '  %s:\n' "${NODE_SERVICE[$node]}"
      printf '    image: "%s"\n' "$IMAGE_NAME"
      printf '    entrypoint:\n'
      printf '      - /usr/bin/gnoland\n'
      printf '    command:\n'
      printf '      - start\n'
      printf '      - -skip-genesis-sig-verification\n'
      printf '      - -data-dir\n'
      printf '      - /data\n'
      printf '      - -genesis\n'
      printf '      - /data/genesis.json\n'
      printf '      - -chainid\n'
      printf '      - %s\n' "$CHAIN_ID"
      printf '      - -gnoroot-dir\n'
      printf '      - /gnoroot\n'
      printf '      - -log-level\n'
      printf '      - %s\n' "$LOG_LEVEL"
      printf '    volumes:\n'
      printf '      - "%s:/data"\n' "${NODE_DATA_DIR[$node]}"
      printf '    ports:\n'
      if [ -n "${NODE_RPC_PORT[$node]:-}" ]; then
        printf '      - "%s:26657"\n' "${NODE_RPC_PORT[$node]}"
      else
        printf '      - "::26657"\n'
      fi
      printf '    networks:\n'
      printf '      - chain\n'
      printf '    stop_grace_period: 5s\n'
    done
    printf '\nnetworks:\n'
    printf '  chain: {}\n'
  } > "$COMPOSE_FILE"
}

create_tx_key() {
  mkdir -p "$KEY_HOME"
  if find "$KEY_HOME" -mindepth 1 -print -quit | grep -q .; then
    return
  fi

  printf '%s\n%s\n%s\n' "$TX_MNEMONIC" "$TX_PASSWORD" "$TX_PASSWORD" | \
    run_gnokey -v "${KEY_HOME}:/keys" \
      add "$TX_KEY_NAME" --home "$(dpath_keys "$KEY_HOME")" --recover --quiet --insecure-password-stdin >/dev/null
}

prepare_network() {
  require_tools
  ensure_runtime_ready

  [ "${#SCENARIO_NODES[@]}" -gt 0 ] || die "no nodes declared"

  rm -rf "$SCENARIO_DIR"
  mkdir -p "$SCENARIO_DIR"

  init_node_dirs
  collect_node_ids
  generate_genesis
  configure_nodes
  [ "$RUNTIME" = "docker" ] && write_compose_file
  create_tx_key

  log "prepared network in ${SCENARIO_DIR}"
}

node_rpc_url() {
  local node="${1:?node required}"
  printf 'http://127.0.0.1:%s' "${NODE_RPC_PORT[$node]}"
}

node_control_url() {
  local node="${1:?node required}"
  [ "${NODE_CONTROLLABLE_SIGNER[$node]:-false}" = "true" ] || die "validator ${node} does not have a controllable signer"
  printf 'http://127.0.0.1:%s' "${NODE_CONTROL_PORT[$node]}"
}

write_inventory() {
  local inventory="${SCENARIO_DIR}/inventory.json"
  local validators_json="[]"
  local node

  for node in "${SCENARIO_VALIDATORS[@]}"; do
    local control_url="null"
    if [ "${NODE_CONTROLLABLE_SIGNER[$node]:-false}" = "true" ]; then
      control_url="\"$(node_control_url "$node")\""
    fi

    validators_json="$(
      jq -cn \
        --argjson current "$validators_json" \
        --arg name "$node" \
        --arg rpc "$(node_rpc_url "$node")" \
        --arg service "${NODE_SERVICE[$node]}" \
        --arg signer_service "${NODE_SIGNER_SERVICE[$node]:-}" \
        --arg address "${NODE_ADDRESS[$node]}" \
        --arg pubkey "${NODE_PUBKEY[$node]}" \
        --argjson controllable "$( [ "${NODE_CONTROLLABLE_SIGNER[$node]:-false}" = "true" ] && printf 'true' || printf 'false' )" \
        --argjson control_url "$control_url" \
        '$current + [{
          name: $name,
          rpc_url: $rpc,
          control_url: $control_url,
          service: $service,
          signer_service: $signer_service,
          controllable_signer: $controllable,
          address: $address,
          pub_key: $pubkey
        }]' \
    )"
  done

  jq -n \
    --arg scenario "$SCENARIO_NAME" \
    --arg work_dir "$SCENARIO_DIR" \
    --arg compose_file "$COMPOSE_FILE" \
    --argjson validators "$validators_json" \
    '{
      scenario: $scenario,
      work_dir: $work_dir,
      compose_file: $compose_file,
      validators: $validators
    }' > "$inventory"

  log "wrote inventory: ${inventory}"
}

wait_for_rpc() {
  local node="${1:?node required}"
  local timeout="${2:-120}"
  local i
  for i in $(seq 1 "$timeout"); do
    if curl -fsS "$(node_rpc_url "$node")/status" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  die "rpc for ${node} did not come up within ${timeout}s"
}

wait_for_control() {
  local node="${1:?node required}"
  local timeout="${2:-120}"
  local i
  for i in $(seq 1 "$timeout"); do
    if curl -fsS "$(node_control_url "$node")/healthz" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  die "control api for ${node} did not come up within ${timeout}s"
}

_capture_node_logs() {
  local node="${1:?node required}"
  # Local processes already redirect their own stdout/stderr to the log file.
  [ "$RUNTIME" = "local" ] && return 0
  # Kill any existing log-follower for this service so there is always exactly
  # one writer per log file (prevents stale followers after container restarts).
  if [ -n "${NODE_LOG_PID[$node]:-}" ]; then
    kill "${NODE_LOG_PID[$node]}" 2>/dev/null || true
  fi
  mkdir -p "${SCENARIO_DIR}/logs"
  # Inline docker compose instead of the compose() wrapper: bash functions are
  # unreliable inside background jobs in non-interactive shells.
  # Pipe through awk to force per-line flushing: docker compose uses full
  # buffering when stdout is not a TTY (i.e. any non-interactive invocation),
  # so without fflush() nothing reaches the log file until the buffer fills.
  # Guard disown with || true — without job control (non-interactive bash)
  # disown can return non-zero which would trigger set -e.
  # Redirect awk stderr to /dev/null so it does not inherit the parent shell's
  # stderr fd — when invoked via runBashScript the parent stderr is a Go pipe,
  # and leaving it open in the background process would cause CombinedOutput()
  # to block indefinitely waiting for the write end to close.
  docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" logs -f "$node" 2>&1 | \
    awk '{ print; fflush() }' >> "${SCENARIO_DIR}/logs/${node}.log" 2>/dev/null &
  local pid="$!"
  NODE_LOG_PID[$node]="$pid"
  disown "$pid" 2>/dev/null || true
}

_resolve_rpc_port() {
  local node="${1:?node required}"
  # Local nodes bind a known deterministic port; nothing to resolve.
  [ "$RUNTIME" = "local" ] && return 0
  local host_port
  host_port="$(compose port "${NODE_SERVICE[$node]}" 26657 2>/dev/null | grep -oE '[0-9]+$')"
  [ -n "$host_port" ] || die "could not resolve host RPC port for ${node}"
  NODE_RPC_PORT[$node]="$host_port"
}

_resolve_control_port() {
  local node="${1:?node required}"
  [ "${NODE_CONTROLLABLE_SIGNER[$node]:-false}" = "true" ] || return 0
  [ "$RUNTIME" = "local" ] && return 0
  local host_port
  host_port="$(compose port "${NODE_SIGNER_SERVICE[$node]}" 8080 2>/dev/null | grep -oE '[0-9]+$')"
  [ -n "$host_port" ] || die "could not resolve host control port for ${node}"
  NODE_CONTROL_PORT[$node]="$host_port"
}

# ---------------------------------------------------------------------------
# Local process supervision. Nodes and signers run as background processes
# whose pids are tracked in NODE_PID / NODE_SIGNER_PID.
# ---------------------------------------------------------------------------

_local_start_node() {
  local node="${1:?node required}"
  local data="${NODE_DATA_DIR[$node]}"
  mkdir -p "${SCENARIO_DIR}/logs"
  "$GNOLAND_BIN" start \
    -skip-genesis-sig-verification \
    -data-dir "$data" \
    -genesis "${data}/genesis.json" \
    -chainid "$CHAIN_ID" \
    -gnoroot-dir "$GNO_ROOT" \
    -log-level "$LOG_LEVEL" \
    >> "${SCENARIO_DIR}/logs/${node}.log" 2>&1 &
  NODE_PID[$node]="$!"
  disown "${NODE_PID[$node]}" 2>/dev/null || true
}

_local_start_signer() {
  local node="${1:?node required}"
  local signer_service="${NODE_SIGNER_SERVICE[$node]}"
  mkdir -p "${SCENARIO_DIR}/logs"
  "$VALSIGNERD_BIN" \
    --key-file "${NODE_DATA_DIR[$node]}/secrets/priv_validator_key.json" \
    --listen-addr "127.0.0.1:${NODE_CONTROL_PORT[$node]}" \
    --remote-signer-addr "tcp://127.0.0.1:${NODE_RS_PORT[$node]}" \
    >> "${SCENARIO_DIR}/logs/${signer_service}.log" 2>&1 &
  NODE_SIGNER_PID[$node]="$!"
  disown "${NODE_SIGNER_PID[$node]}" 2>/dev/null || true
}

# _local_kill_pid PID [GRACE_TENTHS] — SIGTERM then SIGKILL after a grace window.
_local_kill_pid() {
  local pid="${1:-}"
  local grace="${2:-50}"
  [ -n "$pid" ] || return 0
  kill -0 "$pid" 2>/dev/null || return 0
  kill "$pid" 2>/dev/null || true
  local i=0
  while (( i++ < grace )); do
    kill -0 "$pid" 2>/dev/null || return 0
    sleep 0.1
  done
  kill -9 "$pid" 2>/dev/null || true
}

# compose_up_one SERVICE — bring up a single compose service, tolerating Docker's
# transient ephemeral host-port allocation race. The allocator can pick a host
# port that is still briefly held (TIME_WAIT, or not yet released by a
# just-stopped container), making "compose up" fail with "address already in
# use". Each retry removes the just-created container so Docker picks a fresh
# port. Any non-port-race failure aborts immediately so genuine errors surface.
compose_up_one() {
  local service="${1:?service required}"
  local attempt out
  for attempt in 1 2 3; do
    compose rm -fs "$service" >/dev/null 2>&1 || true
    if out="$(compose up -d "$service" 2>&1)"; then
      return 0
    fi
    if printf '%s' "$out" | grep -q "address already in use" && [ "$attempt" -lt 3 ]; then
      log "port bind race starting ${service} on attempt ${attempt}; retrying"
      sleep 1
      continue
    fi
    printf '%s\n' "$out" >&2
    die "failed to start ${service}"
  done
}

start_node() {
  local node="${1:?node required}"
  if [ "$RUNTIME" = "local" ]; then
    _local_start_node "$node"
  else
    # Remove any stopped container so Docker allocates a fresh ephemeral host
    # port rather than reusing the previous binding, which can conflict when
    # multiple nodes are restarted in sequence.
    compose_up_one "$node"
    _resolve_rpc_port "$node"
    _capture_node_logs "$node"
  fi
  wait_for_rpc "$node" 120
  log "started ${node}"
}

start_validator() {
  start_node "$1"
}

start_sentry() {
  start_node "$1"
}

start_all_nodes() {
  [ "${#SCENARIO_NODES[@]}" -gt 0 ] || die "no nodes to start"

  local node

  if [ "${#SCENARIO_SIGNERS[@]}" -gt 0 ]; then
    local signer_service
    for node in "${SCENARIO_SIGNERS[@]}"; do
      signer_service="${NODE_SIGNER_SERVICE[$node]}"
      if [ "$RUNTIME" = "local" ]; then
        _local_start_signer "$node"
      else
        compose_up_one "$signer_service"
        _resolve_control_port "$node"
        _capture_node_logs "$signer_service"
      fi
      wait_for_control "$node" 120
      log "started ${signer_service}"
    done
  fi

  if [ "$RUNTIME" = "local" ]; then
    # Start sentries first (P2P gateway ready), then validators. start_node
    # dispatches to the local backend and waits for RPC; a node's RPC comes up
    # independently of quorum, so per-node waits do not deadlock.
    for node in "${SCENARIO_SENTRIES[@]+"${SCENARIO_SENTRIES[@]}"}"; do
      start_node "$node"
    done
    for node in "${SCENARIO_VALIDATORS[@]+"${SCENARIO_VALIDATORS[@]}"}"; do
      start_node "$node"
    done
    write_inventory
    log "started ${#SCENARIO_NODES[@]} node(s)"
    return 0
  fi

  # Start sentries first and wait for them before launching validators so
  # that the P2P gateway is ready when validators try to dial out.
  if [ "${#SCENARIO_SENTRIES[@]}" -gt 0 ]; then
    local attempt
    for attempt in 1 2 3; do
      if compose up -d "${SCENARIO_SENTRIES[@]}" 2>&1 | grep -q "address already in use"; then
        log "port bind race on attempt ${attempt}; tearing down and retrying"
        compose down --remove-orphans >/dev/null 2>&1 || true
        sleep 1
        continue
      fi
      break
    done
    for node in "${SCENARIO_SENTRIES[@]}"; do
      _resolve_rpc_port "$node"
      _capture_node_logs "$node"
      wait_for_rpc "$node" 120
    done
  fi

  if [ "${#SCENARIO_VALIDATORS[@]}" -gt 0 ]; then
    local attempt
    for attempt in 1 2 3; do
      if compose up -d "${SCENARIO_VALIDATORS[@]}" 2>&1 | grep -q "address already in use"; then
        log "port bind race on attempt ${attempt}; tearing down and retrying"
        compose down --remove-orphans >/dev/null 2>&1 || true
        sleep 1
        continue
      fi
      break
    done
    for node in "${SCENARIO_VALIDATORS[@]}"; do
      _resolve_rpc_port "$node"
      _capture_node_logs "$node"
      wait_for_rpc "$node" 120
    done
  fi

  write_compose_file
  write_inventory
  log "started ${#SCENARIO_NODES[@]} node(s)"
}

stop_node() {
  local node="${1:?node required}"
  if [ "$RUNTIME" = "local" ]; then
    _local_kill_pid "${NODE_PID[$node]:-}"
    NODE_PID[$node]=""
  else
    compose stop "$node" >/dev/null
  fi
  log "stopped ${node}"
}

stop_validator() {
  stop_node "$1"
}

stop_sentry() {
  stop_node "$1"
}

reset_node() {
  local node="${1:?node required}"
  stop_node "$node" || true
  local data="${NODE_DATA_DIR[$node]}"
  if [ "$RUNTIME" = "local" ]; then
    rm -rf "${data}/db" "${data}/wal"
    printf '{"height":"0","round":"0","step":0}\n' > "${data}/secrets/priv_validator_state.json"
    cp "${SCENARIO_DIR}/genesis.json" "${data}/genesis.json"
  else
    # All files under the node data dir are owned by root (created inside the
    # container), so perform the reset from inside a container to avoid host
    # permission errors.
    docker run --rm --entrypoint sh \
      -v "${data}:/data" \
      -v "${SCENARIO_DIR}/genesis.json:/genesis.json:ro" \
      "$IMAGE_NAME" \
      -c 'rm -rf /data/db /data/wal && printf '"'"'{"height":"0","round":"0","step":0}\n'"'"' > /data/secrets/priv_validator_state.json && cp /genesis.json /data/genesis.json'
  fi
  log "reset ${node}"
}

reset_validator() {
  reset_node "$1"
}

safe_reset_node() {
  local node="${1:?node required}"
  stop_node "$node" || true
  # Remove only db and wal; preserve priv_validator_state.json so the node
  # cannot sign a block at a height/round/step it already committed (no double
  # signing). genesis.json is left untouched as well.
  local data="${NODE_DATA_DIR[$node]}"
  if [ "$RUNTIME" = "local" ]; then
    rm -rf "${data}/db" "${data}/wal"
  else
    docker run --rm --entrypoint sh \
      -v "${data}:/data" "$IMAGE_NAME" \
      -c 'rm -rf /data/db /data/wal'
  fi
  log "safe-reset ${node}"
}

safe_reset_validator() {
  safe_reset_node "$1"
}

wait_for_seconds() {
  local seconds="${1:?seconds required}"
  log "waiting ${seconds}s"
  sleep "$seconds"
}

node_height() {
  local node="${1:?node required}"
  curl -fsS "$(node_rpc_url "$node")/status" | jq -r '.result.sync_info.latest_block_height // "0"'
}

wait_for_height() {
  local node="${1:?node required}"
  local target="${2:?target height required}"
  local timeout="${3:-120}"
  local i
  for i in $(seq 1 "$timeout"); do
    local height
    height="$(node_height "$node" 2>/dev/null || printf '0')"
    if [ "$height" -ge "$target" ] 2>/dev/null; then
      log "${node} reached height ${height}"
      return 0
    fi
    sleep 1
  done
  die "${node} did not reach height ${target} within ${timeout}s"
}

wait_for_blocks() {
  local node="${1:?node required}"
  local delta="${2:?delta required}"
  local timeout="${3:-120}"
  local current
  current="$(node_height "$node")"
  wait_for_height "$node" "$((current + delta))" "$timeout"
}

signer_state() {
  local node="${1:?node required}"
  curl -fsS "$(node_control_url "$node")/state"
}

_signer_rule_request() {
  local node="${1:?node required}"
  local phase="${2:?phase required}"
  local action="${3:?action required}"
  local height="${4:-}"
  local round="${5:-}"
  local delay="${6:-}"

  local -a jq_args=(
    -n
    --arg action "$action"
    --arg height "$height"
    --arg round "$round"
    --arg delay "$delay"
  )

  jq "${jq_args[@]}" '
    {
      action: $action
    }
    + (if $height != "" then {height: ($height | tonumber)} else {} end)
    + (if $round != "" then {round: ($round | tonumber)} else {} end)
    + (if $delay != "" then {delay: $delay} else {} end)
  '
}

signer_drop() {
  local node="${1:?validator required}"
  local phase="${2:?phase required}"
  local height="${3:-}"
  local round="${4:-}"
  local payload
  payload="$(_signer_rule_request "$node" "$phase" drop "$height" "$round" "")"
  curl -fsS -X PUT -H 'Content-Type: application/json' --data "$payload" \
    "$(node_control_url "$node")/rules/${phase}" >/dev/null
  log "configured signer drop on ${node} phase=${phase} height=${height:-*} round=${round:-*}"
}

signer_delay() {
  local node="${1:?validator required}"
  local phase="${2:?phase required}"
  local delay="${3:?delay required}"
  local height="${4:-}"
  local round="${5:-}"
  local payload
  payload="$(_signer_rule_request "$node" "$phase" delay "$height" "$round" "$delay")"
  curl -fsS -X PUT -H 'Content-Type: application/json' --data "$payload" \
    "$(node_control_url "$node")/rules/${phase}" >/dev/null
  log "configured signer delay on ${node} phase=${phase} delay=${delay} height=${height:-*} round=${round:-*}"
}

signer_clear() {
  local node="${1:?validator required}"
  local phase="${2:-}"

  if [ -n "$phase" ]; then
    curl -fsS -X DELETE "$(node_control_url "$node")/rules/${phase}" >/dev/null
    log "cleared signer rule on ${node} phase=${phase}"
    return 0
  fi

  curl -fsS -X POST "$(node_control_url "$node")/reset" >/dev/null
  log "cleared signer rules on ${node}"
}

# chain_advances succeeds if the chain produces at least <delta> new blocks on
# <node> within <timeout> seconds. Use this when the caller needs to inspect the
# result before deciding how to fail.
chain_advances() {
  local node="${1:?node required}"
  local timeout="${2:-30}"
  local delta="${3:-2}"
  local before
  before="$(node_height "$node")"
  local target="$((before + delta))"
  local i h
  for i in $(seq 1 "$timeout"); do
    h="$(node_height "$node" 2>/dev/null || printf '0')"
    if [ "$h" -ge "$target" ] 2>/dev/null; then
      log "chain advancing: ${node} reached h=${h} (was ${before})"
      return 0
    fi
    sleep 1
  done
  return 1
}

# assert_chain_halted fails if the chain keeps producing blocks on <node>
# within <timeout> seconds. Use this to verify that a deliberate halt occurred.
assert_chain_halted() {
  local node="${1:?node required}"
  local timeout="${2:-30}"
  local delta="${3:-2}"

  if chain_advances "$node" "$timeout" "$delta"; then
    die "expected chain to halt on ${node}, but it kept advancing"
  fi
  log "chain halted as expected on ${node}"
}

# assert_chain_advances fails if the chain does not produce at least <delta> new
# blocks on <node> within <timeout> seconds. Use this to detect a chain halt.
assert_chain_advances() {
  local node="${1:?node required}"
  local timeout="${2:-30}"
  local delta="${3:-2}"

  if chain_advances "$node" "$timeout" "$delta"; then
    return 0
  fi

  local before
  before="$(node_height "$node" 2>/dev/null || printf '0')"
  local target="$((before + delta))"
  die "chain halted: ${node} height stuck at h=${before} after ${timeout}s (expected >=${target})"
}

docker_network_name() {
  printf '%s' "$NETWORK_NAME"
}

# node_remote NODE — the --remote address gnokey should dial for a node's RPC.
node_remote() {
  local node="${1:?node required}"
  if [ "$RUNTIME" = "local" ]; then
    printf '127.0.0.1:%s' "${NODE_RPC_PORT[$node]}"
  else
    printf '%s:26657' "${NODE_SERVICE[$node]}"
  fi
}

gnokey_tx_with_password() {
  # Consume leading -v <bind> docker volume flags before the gnokey subcommand.
  local -a extra_docker_args=()
  while [[ $# -gt 0 && "$1" == "-v" ]]; do
    extra_docker_args+=("-v" "$2")
    shift 2
  done
  if [ "$RUNTIME" = "local" ]; then
    printf '%s\n' "$TX_PASSWORD" | "$GNOKEY_BIN" "$@"
  else
    printf '%s\n' "$TX_PASSWORD" | \
      docker run -i --rm \
        --entrypoint /usr/bin/gnokey \
        --network "$(docker_network_name)" \
        -v "${KEY_HOME}:/keys" \
        "${extra_docker_args[@]}" \
        "$GNOKEY_IMAGE" \
        "$@"
  fi
}

add_pkg() {
  local target_node="${1:?target node required}"
  local pkgdir="${2:?package dir required}"
  local pkgpath="${3:?package path required}"
  local gas_wanted="${4:-$TX_GAS_WANTED_ADD_PKG}"
  local simulate_mode="${5:-}"

  local abs_pkgdir
  abs_pkgdir="$(cd "$pkgdir" && pwd)"

  local -a cmd=(
    maketx addpkg
    --pkgdir "$(dmount "$abs_pkgdir" /pkg)"
    --pkgpath "$pkgpath"
    --gas-fee "$TX_GAS_FEE"
    --gas-wanted "$gas_wanted"
    --broadcast=true
    --chainid "$CHAIN_ID"
    --remote "$(node_remote "$target_node")"
    --home "$(dpath_keys "$KEY_HOME")"
    --insecure-password-stdin
  )

  if [ -n "$simulate_mode" ]; then
    cmd+=(--simulate "$simulate_mode")
  fi

  cmd+=("$TX_KEY_NAME")

  gnokey_tx_with_password \
    -v "${abs_pkgdir}:/pkg:ro" \
    "${cmd[@]}"
}

estimate_add_pkg_gas() {
  local target_node="${1:?target node required}"
  local pkgdir="${2:?package dir required}"
  local pkgpath="${3:?package path required}"
  local probe_gas_wanted="${4:-$TX_GAS_WANTED_ADD_PKG}"

  local output
  output="$(add_pkg "$target_node" "$pkgdir" "$pkgpath" "$probe_gas_wanted" only)"
  printf '%s\n' "$output" >&2

  local gas_used
  gas_used="$(printf '%s\n' "$output" | awk '/GAS USED:/ {print $3; exit}')"
  [ -n "$gas_used" ] || die "failed to parse simulated gas usage for addpkg on ${target_node}"

  printf '%s\n' "$gas_used"
}

call_realm() {
  local target_node="${1:?target node required}"
  local pkgpath="${2:?package path required}"
  local func_name="${3:?function name required}"
  shift 3 || true

  local -a cmd=(
    maketx call
    --pkgpath "$pkgpath"
    --func "$func_name"
    --gas-fee "$TX_GAS_FEE"
    --gas-wanted "$TX_GAS_WANTED_CALL"
    --broadcast=true
    --chainid "$CHAIN_ID"
    --remote "$(node_remote "$target_node")"
    --home "$(dpath_keys "$KEY_HOME")"
    --insecure-password-stdin
  )

  local arg
  for arg in "$@"; do
    cmd+=(--args "$arg")
  done
  cmd+=("$TX_KEY_NAME")

  gnokey_tx_with_password "${cmd[@]}"
}

run_script() {
  local target_node="${1:?target node required}"
  local script_path="${2:?script path required}"
  local gas_wanted="${3:-$TX_GAS_WANTED_RUN}"
  local simulate_mode="${4:-}"

  local abs_script
  local script_dir
  local script_name
  abs_script="$(cd "$(dirname "$script_path")" && pwd)/$(basename "$script_path")"
  script_dir="$(dirname "$abs_script")"
  script_name="$(basename "$abs_script")"

  local -a cmd=(
    maketx run
      --gas-fee "$TX_GAS_FEE"
      --gas-wanted "$gas_wanted"
      --broadcast=true
      --chainid "$CHAIN_ID"
      --remote "$(node_remote "$target_node")"
      --home "$(dpath_keys "$KEY_HOME")"
      --insecure-password-stdin
  )

  if [ -n "$simulate_mode" ]; then
    cmd+=(--simulate "$simulate_mode")
  fi

  cmd+=("$TX_KEY_NAME" "$(dmount "$abs_script" "/script/${script_name}")")

  gnokey_tx_with_password \
    -v "${script_dir}:/script:ro" \
    "${cmd[@]}"
}

estimate_run_gas() {
  local target_node="${1:?target node required}"
  local script_path="${2:?script path required}"
  local probe_gas_wanted="${3:-$TX_GAS_WANTED_RUN}"

  local output
  output="$(run_script "$target_node" "$script_path" "$probe_gas_wanted" only)"
  printf '%s\n' "$output" >&2

  local gas_used
  gas_used="$(printf '%s\n' "$output" | awk '/GAS USED:/ {print $3; exit}')"
  [ -n "$gas_used" ] || die "failed to parse simulated gas usage for run on ${target_node}"

  printf '%s\n' "$gas_used"
}

send_coins() {
  local target_node="${1:?target node required}"
  local to_addr="${2:?destination address required}"
  local amount="${3:?amount required}"

  gnokey_tx_with_password \
    maketx send \
      --to "$to_addr" \
      --send "$amount" \
      --gas-fee "$TX_GAS_FEE" \
      --gas-wanted "$TX_GAS_WANTED_SEND" \
      --broadcast=true \
      --chainid "$CHAIN_ID" \
      --remote "$(node_remote "$target_node")" \
      --home "$(dpath_keys "$KEY_HOME")" \
      --insecure-password-stdin \
      "$TX_KEY_NAME"
}

do_transaction() {
  local kind="${1:?transaction kind required}"
  shift || true

  case "$kind" in
    addpkg) add_pkg "$@" ;;
    call) call_realm "$@" ;;
    run) run_script "$@" ;;
    send) send_coins "$@" ;;
    *) die "unsupported transaction kind ${kind}" ;;
  esac
}

query_render() {
  local target_node="${1:?target node required}"
  local expr="${2:?render expression required}"

  if [ "$RUNTIME" = "local" ]; then
    "$GNOKEY_BIN" query vm/qrender --data "$expr" --remote "$(node_remote "$target_node")"
  else
    docker run --rm --entrypoint /usr/bin/gnokey --network "$(docker_network_name)" "$GNOKEY_IMAGE" \
      query vm/qrender --data "$expr" --remote "$(node_remote "$target_node")"
  fi
}

container_id_for_node() {
  compose ps -q "$1"
}

node_ip() {
  local node="${1:?node required}"
  if [ "$RUNTIME" = "local" ]; then
    # No container IPs locally; the P2P port stands in as the node's "address".
    printf '%s' "${NODE_P2P_PORT[$node]:-}"
    return 0
  fi
  local container_id
  container_id="$(container_id_for_node "$node")"
  [ -n "$container_id" ] || return 1
  docker inspect "$container_id" | jq -r --arg network "$(docker_network_name)" '.[0].NetworkSettings.Networks[$network].IPAddress // empty'
}

# Local emulation of a sentry IP rotation. There are no container IPs on
# 127.0.0.1, so we stand in by moving the sentry to a new P2P port. Unlike the
# docker scenario (where validators keep dialing a stable DNS name), peers
# address host:port directly, so the nodes that dial this sentry are
# reconfigured and restarted to learn the new port. This exercises sentry
# recreation + reconnection recovery, not the DNS-stable / IP-changed property.
_local_rotate_sentry_ip() {
  local sentry="${1:?sentry name required}"
  local while_down="${2:-}"

  local old_port="${NODE_P2P_PORT[$sentry]}"
  stop_node "$sentry"

  if [ -n "$while_down" ]; then
    "$while_down"
  fi

  local new_port="$((old_port + 1000))"
  NODE_P2P_PORT[$sentry]="$new_port"
  set_config_value "$sentry" p2p.laddr "tcp://127.0.0.1:${new_port}"

  # Find and restart the nodes that dial this sentry so they pick up its new port.
  local node target
  local -a affected=()
  for node in "${SCENARIO_NODES[@]}"; do
    [ "$node" = "$sentry" ] && continue
    while IFS= read -r target; do
      [ "$target" = "$sentry" ] && { affected+=("$node"); break; }
    done < <(persistent_peer_targets "$node")
  done

  for node in "${affected[@]+"${affected[@]}"}"; do
    stop_node "$node"
    apply_peer_config "$node"
  done

  start_node "$sentry"
  for node in "${affected[@]+"${affected[@]}"}"; do
    start_node "$node"
  done

  log "sentry ${sentry} P2P port ${old_port} -> ${new_port} (local emulation)"
}

rotate_sentry_ip() {
  local sentry="${1:?sentry name required}"
  # Optional second argument: name of a shell function to call while the sentry
  # is fully stopped (after removal, before bumpers and restart). Use it to run
  # assertions that require the sentry to be down.
  local while_down="${2:-}"
  [ "${NODE_ROLE[$sentry]:-}" = "sentry" ] || die "${sentry} is not a sentry"

  if [ "$RUNTIME" = "local" ]; then
    _local_rotate_sentry_ip "$sentry" "$while_down"
    return 0
  fi

  local old_ip
  local new_ip
  local bumper
  local bumper2

  old_ip="$(node_ip "$sentry" || true)"
  bumper="${PROJECT_NAME}-${sentry}-bump-1"
  bumper2="${PROJECT_NAME}-${sentry}-bump-2"

  compose stop "$sentry" >/dev/null
  compose rm -f "$sentry" >/dev/null
  docker rm -f "$bumper" "$bumper2" >/dev/null 2>&1 || true

  if [ -n "$while_down" ]; then
    "$while_down"
  fi

  docker run -d --rm --entrypoint sh --name "$bumper" --network "$(docker_network_name)" "$IMAGE_NAME" -c 'sleep 300' >/dev/null
  compose_up_one "$sentry"
  _resolve_rpc_port "$sentry"
  wait_for_rpc "$sentry" 120
  new_ip="$(node_ip "$sentry" || true)"

  if [ -n "$old_ip" ] && [ "$old_ip" = "$new_ip" ]; then
    compose stop "$sentry" >/dev/null
    compose rm -f "$sentry" >/dev/null
    docker run -d --rm --entrypoint sh --name "$bumper2" --network "$(docker_network_name)" "$IMAGE_NAME" -c 'sleep 300' >/dev/null
    compose_up_one "$sentry"
    _resolve_rpc_port "$sentry"
    wait_for_rpc "$sentry" 120
    new_ip="$(node_ip "$sentry" || true)"
  fi

  docker rm -f "$bumper" "$bumper2" >/dev/null 2>&1 || true
  [ -n "$new_ip" ] || die "failed to resolve a new IP for sentry ${sentry}"
  if [ -n "$old_ip" ] && [ "$old_ip" = "$new_ip" ]; then
    die "sentry ${sentry} kept IP ${new_ip} after recreation; rotation scenario was not exercised"
  fi
  log "sentry ${sentry} IP ${old_ip:-unknown} -> ${new_ip:-unknown}"
}

print_cluster_status() {
  local node
  for node in "${SCENARIO_NODES[@]}"; do
    if curl -fsS "$(node_rpc_url "$node")/status" >/dev/null 2>&1; then
      printf '%-16s role=%-10s height=%s rpc=%s\n' \
        "$node" \
        "${NODE_ROLE[$node]}" \
        "$(node_height "$node")" \
        "$(node_rpc_url "$node")"
    else
      printf '%-16s role=%-10s state=stopped rpc=%s\n' \
        "$node" \
        "${NODE_ROLE[$node]}" \
        "$(node_rpc_url "$node")"
    fi
  done
}

scenario_finish() {
  if [ "$RUNTIME" = "local" ]; then
    if [ "${KEEP_UP:-0}" = "1" ]; then
      log "leaving processes running because KEEP_UP=1"
      return 0
    fi
    local node
    for node in "${SCENARIO_NODES[@]+"${SCENARIO_NODES[@]}"}"; do
      _local_kill_pid "${NODE_PID[$node]:-}"
      _local_kill_pid "${NODE_SIGNER_PID[$node]:-}"
    done
    return 0
  fi

  local sentry
  for sentry in "${SCENARIO_SENTRIES[@]+"${SCENARIO_SENTRIES[@]}"}"; do
    docker rm -f "${PROJECT_NAME}-${sentry}-bump-1" "${PROJECT_NAME}-${sentry}-bump-2" >/dev/null 2>&1 || true
  done
  if [ "${KEEP_UP:-0}" = "1" ]; then
    log "leaving network running because KEEP_UP=1"
    return 0
  fi
  if [ -f "$COMPOSE_FILE" ]; then
    compose down --remove-orphans >/dev/null 2>&1 || true
  fi
}
