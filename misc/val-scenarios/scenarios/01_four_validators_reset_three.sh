#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=true

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/lib/scenario.sh"

scenario_init "scenario-01"
trap scenario_finish EXIT

gen_validator val1
gen_validator val2
gen_validator val3
gen_validator val4

prepare_network
start_all_nodes
assert_chain_advances val1 120 5

stop_validator val2
stop_validator val3
stop_validator val4

# 1/4 validators running — chain must halt.
assert_chain_halted val1 30

reset_validator val2
reset_validator val3
reset_validator val4

start_validator val2
start_validator val3
start_validator val4

# BUG: once block sync works correctly for reset validators, the chain should
# resume after all 4 validators restart. On unpatched master, the reset
# validators cannot complete block sync and rejoin consensus, so the chain
# remains halted. When fixed, replace the assertion below with:
#   assert_chain_advances val1 120 2
assert_chain_halted val1 30

print_cluster_status
