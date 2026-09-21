#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=true

# 4 validators with unequal voting power: 10 / 1 / 1 / 1 (total = 13).
# The 2/3 threshold is ceil(2/3 * 13) = 9.
# val1 alone holds 10 > 9, so stopping val2, val3, and val4 must not halt
# the chain.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/lib/scenario.sh"

scenario_init "scenario-11"
trap scenario_finish EXIT

gen_validator val1 --power 10
gen_validator val2
gen_validator val3
gen_validator val4

prepare_network
start_all_nodes
assert_chain_advances val1 120 5

stop_validator val2
stop_validator val3
stop_validator val4

# val1 holds 10/13 (≈77%) of voting power — chain must keep advancing.
assert_chain_advances val1 60 2

print_cluster_status
