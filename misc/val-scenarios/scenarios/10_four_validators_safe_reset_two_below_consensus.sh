#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=true
# Realistic canary: unlike the other consensus-only scenarios, this one KEEPS the
# full example genesis (SCENARIO_GENESIS_EXAMPLES defaults to true). It is the one
# CI scenario that exercises a halt below quorum and resume while validators must
# replay a realistic full genesis at InitChain on restart. Do not set
# SCENARIO_GENESIS_EXAMPLES=false here — that coverage would otherwise be lost.

# 4 validators, safe reset 2 (db + wal only, priv_validator_state preserved).
# 2/4 remain during the reset (50% < 2/3 threshold) so the chain must halt.
# After the two validators are restarted the chain must resume.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/lib/scenario.sh"

scenario_init "scenario-10"
trap scenario_finish EXIT

gen_validator val1
gen_validator val2
gen_validator val3
gen_validator val4

prepare_network
start_all_nodes
assert_chain_advances val1 120 5

safe_reset_validator val3
safe_reset_validator val4

# 2/4 validators running — chain must halt.
# delta=4: absorbs up to 2 in-flight blocks that val3/val4 may have
# pre-committed before SIGTERM before the chain truly halts.
assert_chain_halted val1 30 4

start_validator val3
start_validator val4

# The chain must resume once 4/4 validators are running again.
assert_chain_advances val1 120 2

# val3 and val4 must catch up to the current chain height via block sync, then
# actively produce new blocks (proving they re-joined consensus).
sync_target="$(node_height val1)"
wait_for_height val3 "$sync_target" 120
wait_for_height val4 "$sync_target" 120
assert_chain_advances val3 60 2
assert_chain_advances val4 60 2

print_cluster_status
