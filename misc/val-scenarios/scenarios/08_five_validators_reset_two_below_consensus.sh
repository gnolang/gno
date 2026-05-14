#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=true

# 5 validators, stop/reset/restart 2.
# 3/5 remain during the reset (60% < 2/3 threshold) so the chain must halt.
# After the two validators are restarted the chain must resume.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/lib/scenario.sh"

scenario_init "scenario-08"
trap scenario_finish EXIT

gen_validator val1
gen_validator val2
gen_validator val3
gen_validator val4
gen_validator val5

prepare_network
start_all_nodes
assert_chain_advances val1 120 5

stop_validator val4
stop_validator val5
reset_validator val4
reset_validator val5

# 3/5 validators running — chain must halt.
assert_chain_halted val1 30

start_validator val4
start_validator val5

# BUG: once block sync works correctly for reset validators, the chain should
# resume and val4/val5 should catch up and actively produce new blocks. On
# unpatched master, the reset validators cannot complete block sync and rejoin
# consensus, so the chain remains halted. When fixed, replace the assertion
# below with:
#   assert_chain_advances val1 120 2
#   sync_target="$(node_height val1)"
#   wait_for_height val4 "$sync_target" 120
#   wait_for_height val5 "$sync_target" 120
#   assert_chain_advances val4 60 2
#   assert_chain_advances val5 60 2
assert_chain_halted val1 30

print_cluster_status
