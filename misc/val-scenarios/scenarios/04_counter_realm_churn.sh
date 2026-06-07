#!/usr/bin/env bash
set -euo pipefail

SCENARIO_CI=true

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/lib/scenario.sh"

scenario_init "scenario-04"
trap scenario_finish EXIT

gen_validator val1
gen_validator val2
gen_validator val3
gen_validator val4

prepare_network
start_all_nodes
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

add_pkg val1 "$pkg_dir" "gno.land/r/demo/scenario_counter"
wait_for_seconds 5

call_realm val1 "gno.land/r/demo/scenario_counter" "Increment"
call_realm val1 "gno.land/r/demo/scenario_counter" "Increment"
call_realm val1 "gno.land/r/demo/scenario_counter" "Increment"
wait_for_seconds 10

stop_validator val3
reset_validator val3
start_validator val3
wait_for_seconds 20

call_realm val1 "gno.land/r/demo/scenario_counter" "Increment"

result="$(query_render val1 "gno.land/r/demo/scenario_counter:" | awk '/^data:/ {print $2}')"
[ "$result" = "4" ] || die "expected counter=4, got: ${result}"

print_cluster_status
