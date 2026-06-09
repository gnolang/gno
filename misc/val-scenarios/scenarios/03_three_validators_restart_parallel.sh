#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=true

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/lib/scenario.sh"

scenario_init "scenario-03"
trap scenario_finish EXIT

gen_validator val1
gen_validator val2
gen_validator val3

prepare_network
start_all_nodes
assert_chain_advances val1 120 5

stop_validator val1
stop_validator val2
stop_validator val3
wait_for_seconds 5

start_validator val1
start_validator val2
start_validator val3
assert_chain_advances val1 120 2

print_cluster_status
