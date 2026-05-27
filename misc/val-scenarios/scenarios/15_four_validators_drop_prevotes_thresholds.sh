#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=false

# 4 validators with controllable signer sidecars.
# First, 1/4 validators drops prevotes, leaving 3/4 prevoting (> 2/3), so the
# chain must keep advancing.
# Then, 3/4 validators drop prevotes, leaving only 1/4 prevoting (< 2/3), so
# the chain must halt.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/lib/scenario.sh"

scenario_init "scenario-15"
trap scenario_finish EXIT

gen_validator val1 --controllable-signer
gen_validator val2 --controllable-signer
gen_validator val3 --controllable-signer
gen_validator val4 --controllable-signer

prepare_network
start_all_nodes
assert_chain_advances val1 120 5

signer_drop val4 prevote

# 3/4 validators still sign prevotes, which stays above the 2/3 threshold.
assert_chain_advances val1 60 2

signer_drop val2 prevote
signer_drop val3 prevote

# Now only val1 can prevote. With 1/4 signing prevotes, no proposal can gather
# a quorum and the chain must halt.
assert_chain_halted val1 30

print_cluster_status
