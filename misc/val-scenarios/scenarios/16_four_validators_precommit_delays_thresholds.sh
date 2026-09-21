#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=false

# Keep the commit timeout fixed so the delay thresholds below stay meaningful.
TIMEOUT_COMMIT="1s"

# 4 validators with controllable signer sidecars.
# First validator:
# - delay precommits below timeout_commit and assert the chain keeps advancing
# - increase the delay above timeout_commit and assert the chain still advances
# Second validator:
# - repeat the same under/over-timeout progression and assert the chain keeps
#   advancing while at least one delayed precommit still arrives quickly enough
# Finally:
# - increase both delayed validators well beyond the observation window and
#   assert that no blocks are committed for the next 10 seconds

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/lib/scenario.sh"

scenario_init "scenario-16"
trap scenario_finish EXIT

gen_validator val1 --controllable-signer
gen_validator val2 --controllable-signer
gen_validator val3 --controllable-signer
gen_validator val4 --controllable-signer

prepare_network
start_all_nodes
assert_chain_advances val1 120 5

signer_delay val4 precommit 500ms

# 3/4 validators still precommit immediately and val4 still signs before the
# 1s timeout_commit elapses, so blocks should keep finalizing normally.
assert_chain_advances val1 60 2

signer_delay val4 precommit 1500ms

# val4 now signs after timeout_commit, but the other 3 validators still provide
# a quorum of precommits, so consensus must keep moving.
assert_chain_advances val1 60 2

signer_delay val3 precommit 500ms

# val3 remains below timeout_commit while val4 is slow, so the chain should
# still finalize once val3's precommit arrives.
assert_chain_advances val1 60 2

signer_delay val3 precommit 1500ms

# Both delayed validators are now beyond timeout_commit, but either delayed
# precommit can still complete the 3/4 quorum after ~1.5s, so the chain must
# continue, albeit more slowly.
assert_chain_advances val1 60 2

signer_delay val3 precommit 20s
signer_delay val4 precommit 20s

# With both delayed precommits pushed past the next 10s observation window,
# only 2/4 validators can precommit in time, so no new block should commit.
assert_chain_halted val1 10 1

print_cluster_status
