#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=true
# Consensus-only: skip example packages + on-chain PoA valset realm in genesis
# (validators reach consensus via the genesis validator set).
SCENARIO_GENESIS_EXAMPLES=false

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

# After reset validators restart they complete block sync and rejoin consensus;
# the chain resumes once 3/4 validators are back.
assert_chain_advances val1 120 2

print_cluster_status
