#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=true

# Check whether a single image binary is affected by the gas non-determinism
# bug fixed by PR #5400.
#
# Root cause: with CacheStdlibLoad=false, genesis-fresh nodes populate stdlib
# into a per-tx clone during LoadStdlib and discard it, so vm.typeCheckCache
# stays empty. Restarted nodes run Initialize and do populate vm.typeCheckCache.
# When a package that imports strconv is deployed with a tight gas budget, the
# cold nodes charge much more gas than the warm nodes, causing consensus
# divergence and a chain halt.
#
# This scenario uses one binary only:
# - start a fresh 4-validator network
# - restart val3/val4 so they become "warm"
# - estimate addpkg gas on a restarted "warm" validator
# - submit the same addpkg with that gas limit and --simulate=skip so cold
#   validators cannot reject it locally before consensus sees it
# If the chain keeps advancing after deployment the binary is clean; if it
# halts, the binary is affected by the bug.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/lib/scenario.sh"

ADD_PKG_PROBE_GAS_WANTED="${ADD_PKG_PROBE_GAS_WANTED:-50000000}"
ADD_PKG_GAS_MARGIN="${ADD_PKG_GAS_MARGIN:-0}"

scenario_init "scenario-06"
trap scenario_finish EXIT

gen_validator val1
gen_validator val2
gen_validator val3
gen_validator val4

prepare_network
start_all_nodes

# Let the cluster reach a stable height before restarting a subset.
wait_for_blocks val1 5 120

# Restart val3 and val4 with existing data (no reset).
# On buggy binaries, this is what makes val3/val4 warm while val1/val2 remain
# cold from the original genesis start.
restart_height="$(node_height val1)"
stop_validator val3
stop_validator val4
start_validator val3
start_validator val4

# Wait for the restarted validators to catch up, then let the cluster produce a
# few more blocks before the trigger tx.
wait_for_height val3 "${restart_height}" 120
wait_for_height val4 "${restart_height}" 120
wait_for_blocks val1 5 120

# Generate the counter realm package inline so no external packages/ dir is needed.
pkg_dir="${SCENARIO_DIR}/packages/scenario-counter"
mkdir -p "$pkg_dir"

cat > "${pkg_dir}/gnomod.toml" << 'EOF'
module = "gno.land/r/demo/scenario_counter"
gno = "0.9"
EOF

cat > "${pkg_dir}/counter.gno" << 'EOF'
package scenario_counter

import "strconv"

var counter int

func Increment(_ realm) int {
	counter++
	return counter
}

func Render(_ string) string {
	return strconv.Itoa(counter)
}
EOF

# Estimate the addpkg gas on a restarted warm validator, then submit the same
# tx through that warm validator with --simulate=skip. On buggy binaries this
# lets the tx reach consensus while still being too expensive for the cold
# validators.
warm_gas_used="$(
  estimate_add_pkg_gas \
    val3 \
    "$pkg_dir" \
    "gno.land/r/demo/scenario_counter" \
    "${ADD_PKG_PROBE_GAS_WANTED}"
)"
add_pkg_gas_wanted="$((warm_gas_used + ADD_PKG_GAS_MARGIN))"
log "warm addpkg gas estimate from val3: ${warm_gas_used}; using gas-wanted=${add_pkg_gas_wanted}"

# Buggy binaries: cold nodes need more gas than the warm estimate, so consensus
# diverges and the chain halts. Fixed binaries: all nodes compute the same gas
# and the chain continues.
#
# The submit RPC may report either OK or OOG depending on which validator
# executed the tx locally. The real signal is whether the cluster keeps
# producing blocks afterwards.
set +e
tx_output="$(
  add_pkg \
    val3 \
    "$pkg_dir" \
    "gno.land/r/demo/scenario_counter" \
    "${add_pkg_gas_wanted}" \
    skip 2>&1
)"
tx_status=$?
set -e
printf '%s\n' "$tx_output"
if [ "$tx_status" -ne 0 ]; then
  log "trigger tx returned non-zero from val3; checking whether the cluster halted"
fi

# Verify whether the chain keeps producing blocks after the deployment.
# If it halts, the image is affected by the bug.
if chain_advances val1 30 2; then
  print_cluster_status
else
  print_cluster_status
  log "chain halted after trigger tx"
  exit 1
fi
