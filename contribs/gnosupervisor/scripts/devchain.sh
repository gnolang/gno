#!/usr/bin/env bash
# A local single-validator gno.land chain that halts through a real GovDAO
# proposal (node:p:halt_height and node:p:halt_min_version, set with
# misc/govdao-scripts/set-halt.sh), then runs a newer gnoland after the halt.
# It is the chain internal/supervisor/realchain_test.go replays from genesis.
#
# Usage: devchain.sh <command>
#   build   build gnoland OLD and NEW (they differ only by version) and gnokey
#   up      start node A on OLD, halt it by GovDAO vote at H, restart it on NEW
#   env     write and print run/nodeb.env, the input of realchain_test.go
#   status  print node A's pid and height
#   down    stop the processes this script started; --purge also removes them
#
# Everything is written under $DEVCHAIN_DIR (default: gnosupervisor-devchain-<uid>
# in $TMPDIR). Native processes only, on ports 36656/36657 by
# default; no tracked file is modified.
set -euo pipefail

GNOROOT="${GNOROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)}"
export GNOROOT # -trimpath binaries cannot guess it outside the tree
R="${DEVCHAIN_DIR:-${TMPDIR:-/tmp}/gnosupervisor-devchain-$(id -u)}"
# This script runs the binaries it finds under $R: refuse a directory someone
# else could have prepared.
if [ -e "$R" ] && { [ -L "$R" ] || [ ! -O "$R" ]; }; then
	echo "error: $R is a symlink or is not owned by $(id -un)" >&2
	exit 1
fi
mkdir -p -m 700 "$R"
BIN="$R/bin"
RUN="$R/run"
OLD="${OLD_VERSION:-v9.0.0}"
NEW="${NEW_VERSION:-v9.1.0}"
CHAIN_ID=dev # r/gov/dao/init/v0 only works on chain "dev"
A_RPC="${A_RPC:-36657}" A_P2P="${A_P2P:-36656}"
HALT_OFFSET="${HALT_OFFSET:-15}"
POST_HALT_BLOCKS="${POST_HALT_BLOCKS:-10}"
TIMEOUT_COMMIT="${TIMEOUT_COMMIT:-1s}"
TEST1_ADDR=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
# Public test mnemonic shipped in the repo (gno.land/pkg/integration/node_testing.go).
TEST1_MNEMONIC="source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"

A_DIR="$RUN/nodeA" GENESIS="$RUN/genesis.json"
KEYS="$RUN/gnokey-home" LOGS="$RUN/logs"

log() { printf '[%s] %s\n' "$(date +%H:%M:%S)" "$*" >&2; }
die() { log "ERROR: $*"; exit 1; }

# ---------------------------------------------------------------- build
cmd_build() {
	mkdir -p "$BIN"
	local t0=$SECONDS v
	for v in "$OLD" "$NEW"; do
		(cd "$GNOROOT" && CGO_ENABLED=0 go build -trimpath \
			-ldflags "-X github.com/gnolang/gno/tm2/pkg/version.Version=$v" \
			-o "$BIN/gnoland-$v" ./gno.land/cmd/gnoland)
	done
	(cd "$GNOROOT" && CGO_ENABLED=0 go build -trimpath -o "$BIN/gnokey" ./gno.land/cmd/gnokey)
	"$BIN/gnoland-$OLD" version
	"$BIN/gnoland-$NEW" version
	log "build done in $((SECONDS - t0))s"
}

# ---------------------------------------------------------------- helpers
port_free() { ! ss -ltnH "sport = :$1" | grep -q .; }

height() { # height RPCPORT -> latest_block_height or empty
	curl -fsS --max-time 2 "http://127.0.0.1:$1/status" 2>/dev/null |
		jq -r '.result.sync_info.latest_block_height' 2>/dev/null || true
}

wait_rpc() { # wait_rpc PORT TIMEOUT PID
	local i
	for ((i = 0; i < $2; i++)); do
		[ -n "$(height "$1")" ] && return 0
		kill -0 "$3" 2>/dev/null || return 1
		sleep 1
	done
	return 1
}

wait_height() { # wait_height PORT TARGET TIMEOUT
	local i h
	for ((i = 0; i < $3; i++)); do
		h=$(height "$1")
		[ -n "$h" ] && [ "$h" -ge "$2" ] && return 0
		sleep 1
	done
	return 1
}

init_node() { # init_node DIR RPC P2P
	local d=$1 g="$BIN/gnoland-$OLD"
	mkdir -p "$d"
	"$g" config init --config-path "$d/config/config.toml" >/dev/null
	"$g" secrets init --data-dir "$d/secrets" >/dev/null
	local c=(--config-path "$d/config/config.toml")
	"$g" config set "${c[@]}" rpc.laddr "tcp://127.0.0.1:$2" >/dev/null
	"$g" config set "${c[@]}" p2p.laddr "tcp://127.0.0.1:$3" >/dev/null
	"$g" config set "${c[@]}" p2p.pex false >/dev/null
	"$g" config set "${c[@]}" p2p.allow_duplicate_ip true >/dev/null
	"$g" config set "${c[@]}" consensus.timeout_commit "$TIMEOUT_COMMIT" >/dev/null
	# Must stay 0: a config halt_height would mix with the governance one.
	[ "$("$g" config get "${c[@]}" halt_height --raw 2>/dev/null || echo 0)" = 0 ] || die "halt_height set in $d config"
}

start_node() { # start_node NAME VERSION DIR [extra flags...]  -> sets PID
	local name=$1 ver=$2 dir=$3
	shift 3
	mkdir -p "$LOGS"
	"$BIN/gnoland-$ver" start \
		--data-dir "$dir" --genesis "$GENESIS" --chainid "$CHAIN_ID" \
		--gnoroot-dir "$GNOROOT" --skip-genesis-sig-verification --log-level info "$@" \
		>>"$LOGS/$name.log" 2>&1 &
	PID=$!
	echo "$PID" >"$RUN/$name.pid"
	log "started $name ($ver) pid=$PID log=$LOGS/$name.log"
}

stop_pidfile() { # stop_pidfile NAME
	local f="$RUN/$1.pid" pid i
	[ -f "$f" ] || return 0
	pid=$(cat "$f")
	# Only kill it if it is still one of our binaries.
	if kill -0 "$pid" 2>/dev/null && tr '\0' ' ' <"/proc/$pid/cmdline" | grep -q "^$BIN/gnoland-"; then
		kill -TERM "$pid"
		for ((i = 0; i < 30; i++)); do kill -0 "$pid" 2>/dev/null || break; sleep 1; done
		kill -0 "$pid" 2>/dev/null && kill -KILL "$pid"
		log "stopped $1 (pid $pid)"
	fi
	rm -f "$f"
}

gk() { GNOHOME="$KEYS" "$BIN/gnokey" "$@"; }

tx() { # tx <maketx args...> ; empty password on stdin
	echo "" | GNOHOME="$KEYS" "$BIN/gnokey" maketx "$1" -insecure-password-stdin \
		-gas-fee 10000000ugnot -gas-wanted 100000000 -broadcast \
		-chainid "$CHAIN_ID" -remote "tcp://127.0.0.1:$A_RPC" "${@:2}"
}

qparam() { gk query "params/$1" -remote "tcp://127.0.0.1:$A_RPC" 2>/dev/null | sed -n 's/^data: //p'; }

# ---------------------------------------------------------------- up
cmd_up() {
	[ -x "$BIN/gnoland-$OLD" ] && [ -x "$BIN/gnoland-$NEW" ] || cmd_build
	for p in "$A_RPC" "$A_P2P"; do port_free "$p" || die "port $p busy"; done
	[ ! -e "$RUN" ] || die "$RUN exists; run '$0 cleanup' first"
	mkdir -p "$RUN" "$LOGS"

	# gnokey home with test1 (GovDAO member once InitWithUsers ran).
	printf '%s\n\n\n' "$TEST1_MNEMONIC" | gk add test1 -recover -insecure-password-stdin >/dev/null 2>&1
	# set-halt.sh calls a bare `gnokey maketx run ...` that prompts on a TTY:
	# shim it to use our home and an empty password from stdin.
	mkdir -p "$RUN/shim"
	cat >"$RUN/shim/gnokey" <<EOF
#!/usr/bin/env bash
if [ "\$1" = maketx ]; then
  sub=\$2; shift 2
  echo "" | GNOHOME="$KEYS" exec "$BIN/gnokey" maketx "\$sub" -insecure-password-stdin "\$@"
  exit \$?
fi
GNOHOME="$KEYS" exec "$BIN/gnokey" "\$@"
EOF
	chmod +x "$RUN/shim/gnokey"

	init_node "$A_DIR" "$A_RPC" "$A_P2P"

	local t0=$SECONDS
	start_node nodeA "$OLD" "$A_DIR" --lazy
	wait_rpc "$A_RPC" 300 "$PID" || die "node A did not come up; see $LOGS/nodeA.log"
	wait_height "$A_RPC" 2 120 || die "node A produces no blocks"
	log "node A up at height $(height "$A_RPC") after $((SECONDS - t0))s"

	# 1. Make test1 the single T1 GovDAO member (dev-chain-only helper).
	cat >"$RUN/init_govdao.gno" <<EOF
package main

import dao "gno.land/r/gov/dao/init/v0"

func main(cur realm) {
	dao.InitWithUsers(cross(cur), "$TEST1_ADDR")
}
EOF
	tx run test1 "$RUN/init_govdao.gno" >"$LOGS/tx-init.log" 2>&1 || { cat "$LOGS/tx-init.log"; die "InitWithUsers failed"; }
	log "GovDAO initialised with test1"

	# 2. Create the halt proposal with the repo script (real GovDAO path).
	local cur H
	cur=$(height "$A_RPC")
	H=$((cur + HALT_OFFSET))
	echo "$H" >"$RUN/halt_height"
	PATH="$RUN/shim:$PATH" GNOKEY_NAME=test1 CHAIN_ID="$CHAIN_ID" \
		REMOTE="127.0.0.1:$A_RPC" GAS_WANTED=100000000 GAS_FEE=10000000ugnot \
		"$GNOROOT/misc/govdao-scripts/set-halt.sh" "$H" "$NEW" >"$LOGS/tx-set-halt.log" 2>&1 ||
		{ cat "$LOGS/tx-set-halt.log"; die "set-halt.sh failed"; }
	log "halt proposal created: H=$H min_version=$NEW (proposed at height $cur)"

	# InitWithUsers installs a fresh GovDAO instance, so this is proposal 0.
	local pid_prop=0
	gk query vm/qrender -remote "tcp://127.0.0.1:$A_RPC" -data "gno.land/r/gov/dao:$pid_prop" 2>/dev/null |
		grep -q "halt" || die "proposal $pid_prop is not the halt proposal"
	tx call -pkgpath gno.land/r/gov/dao -func MustVoteOnProposalSimple -args "$pid_prop" -args YES test1 \
		>"$LOGS/tx-vote.log" 2>&1 || { cat "$LOGS/tx-vote.log"; die "vote failed"; }
	tx call -pkgpath gno.land/r/gov/dao -func ExecuteProposal -args "$pid_prop" test1 \
		>"$LOGS/tx-exec.log" 2>&1 || { cat "$LOGS/tx-exec.log"; die "execute failed"; }
	log "proposal $pid_prop voted + executed at height $(height "$A_RPC")"
	log "params: halt_height=$(qparam node:p:halt_height) halt_min_version=$(qparam node:p:halt_min_version)"

	# 3. Wait for the halt: height reaches H then stops moving.
	local t1=$SECONDS
	wait_height "$A_RPC" "$H" $((HALT_OFFSET * 10 + 60)) || die "node A never reached H=$H"
	local last="" same=0 h
	while [ $same -lt 8 ]; do
		sleep 1
		h=$(height "$A_RPC")
		if [ "$h" = "$last" ]; then same=$((same + 1)); else same=0; last=$h; fi
	done
	kill -0 "$(cat "$RUN/nodeA.pid")" || die "node A process exited (unexpected at the tip)"
	log "node A halted: /status height=$last (H=$H), process still up; reached H $((SECONDS - t1))s after execute"
	grep -E "GovDAO halt height reached|CONSENSUS FAILURE" "$LOGS/nodeA.log" | cut -c1-220 | head -3 >&2 || true

	# 4. Upgrade: stop A, restart on NEW, same data dir.
	stop_pidfile nodeA
	echo "---- restart on $NEW ----" >>"$LOGS/nodeA.log"
	start_node nodeA "$NEW" "$A_DIR"
	wait_rpc "$A_RPC" 300 "$PID" || die "node A ($NEW) did not come up; see $LOGS/nodeA.log"
	wait_height "$A_RPC" $((H + POST_HALT_BLOCKS)) 120 || die "node A ($NEW) did not pass H"
	log "node A on $NEW at height $(height "$A_RPC") > H=$H"
	cmd_env
}

cmd_start_a() {
	port_free "$A_RPC" || die "port $A_RPC busy (node A already running?)"
	start_node nodeA "$NEW" "$A_DIR"
	wait_rpc "$A_RPC" 300 "$PID" || die "node A did not come up"
}

# ---------------------------------------------------------------- env
cmd_env() {
	local id H
	id=$(curl -fsS "http://127.0.0.1:$A_RPC/status" | jq -r '.result.node_info.net_address | split("@")[0]')
	H=$(cat "$RUN/halt_height")
	cat >"$RUN/nodeb.env" <<EOF
# What a node syncing from genesis across the halt needs (see realchain_test.go).
GNOROOT=$GNOROOT
CHAIN_ID=$CHAIN_ID
GENESIS=$GENESIS
HALT_HEIGHT=$H
HALT_MIN_VERSION=$NEW
OLD_BIN=$BIN/gnoland-$OLD
NEW_BIN=$BIN/gnoland-$NEW
NODE_A_ID=$id
NODE_A_P2P=$id@127.0.0.1:$A_P2P
NODE_A_RPC=http://127.0.0.1:$A_RPC
EOF
	cat "$RUN/nodeb.env"
}

cmd_status() {
	[ -f "$RUN/nodeA.pid" ] && echo "nodeA pid=$(cat "$RUN/nodeA.pid") alive=$(kill -0 "$(cat "$RUN/nodeA.pid")" 2>/dev/null && echo yes || echo no)"
	echo "A height=$(height "$A_RPC") H=$(cat "$RUN/halt_height" 2>/dev/null)"
}

cmd_down() {
	stop_pidfile nodeA
	if [ "${1:-}" = "--purge" ]; then rm -rf "$R"; log "removed $R"; fi
}

case "${1:-}" in
build) cmd_build ;;
up) cmd_up ;;
env) cmd_env ;;
status) cmd_status ;;
down) shift; cmd_down "$@" ;;
*) sed -n '2,16p' "$0"; exit 1 ;;
esac
