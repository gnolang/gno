#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=false

# 4 validators with controllable signer sidecars.
# Drop proposal signatures on all validators while they stay online.
# Consensus keeps advancing rounds via timeout_propose, but no blocks can be
# committed, so the chain must halt at a fixed height.
# Clearing the signer rules should let the chain resume without restarting
# nodes, because proposal signing is retried at each new round.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/lib/scenario.sh"

scenario_init "scenario-14"
trap scenario_finish EXIT

gen_validator val1 --controllable-signer
gen_validator val2 --controllable-signer
gen_validator val3 --controllable-signer
gen_validator val4 --controllable-signer

prepare_network
start_all_nodes
assert_chain_advances val1 120 5

for validator in val1 val2 val3 val4; do
  signer_drop "$validator" proposal
done

# Validators stay online and keep timing out into new rounds, but no proposer
# can sign a proposal, so commits must stop.
assert_chain_halted val1 30

for validator in val1 val2 val3 val4; do
  signer_clear "$validator" proposal
done

# No reset/restart needed: once proposers can sign again, the next successful
# round should recover consensus and commit new blocks again.
assert_chain_advances val1 120 2
assert_chain_advances val4 120 2

print_cluster_status
